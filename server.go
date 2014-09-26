package main
import(
	"os"
	"io"
	"log"
	"time"
	"fmt"
	"strings"
	"strconv"
	"flag"
	"os/exec"
	"math/rand"
	"net/http"
	"encoding/json"
	"github.com/gorilla/mux"
//	"github.com/gorilla/sessions"
	"github.com/boj/redistore"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	
)


var (
	dbsession *mgo.Session
	store *redistore.RediStore
	SERVER_PORT = flag.Int("port",7313,"port to serve on")
)

const (
	DBNAME = "glass_alerts"
	COLNAME = "DBSyncHelper"
	SESSION_NAME = "deere_planting_server"
	ACCOUNT_SESSION_KEY = "account_name"
	DUMMY_PASSWORD = "nothingrunslikeadeere"

	BASE_PIC_DIR = "/usr/share/repco/GlassDeereBackend/"
)
func httpError(w http.ResponseWriter,error string,code int){
	log.Printf("httpError %d : %s",code,error)
	http.Error(w,error,code)
}
func main(){
	flag.Parse()

	var err error

	//initialize database
	dbsession,err = mgo.Dial("localhost")
	if err != nil{
		log.Fatalf("mongodb error: %s",err.Error())
	}

	defer dbsession.Close()
	
	dbsession.DB(DBNAME).C(COLNAME).EnsureIndexKey("alert_data.ts_millis")
	dbsession.DB(DBNAME).C(COLNAME).EnsureIndexKey("account_name")
	
	//initialize session store 

	//TODO: export password to config file outside source tree
	store,err = redistore.NewRediStore(10,"tcp",":6379","",[]byte("supersecretpassword"))
	if err != nil{
		log.Fatalf("redistore error: %s",err.Error())
	}
	defer store.Close()
	

	//initialize routing
	r := mux.NewRouter()
	r.HandleFunc("/session",sessionStatus).Methods("GET")
	r.HandleFunc("/alerts",getAlerts).Methods("GET")
	r.HandleFunc("/sync",SyncHandler).Methods("POST")
	r.HandleFunc("/logout",logoutSession).Methods("GET")
	r.HandleFunc("/login",login).Methods("POST")

	http.Handle("/",r)

	log.Printf("serving on port %d",*SERVER_PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d",*SERVER_PORT),nil))
}

func getAlertsByAccount(accountName string) *mgo.Query{
	filter := bson.M{
		"account_name" : accountName,
	}
	return dbsession.DB(DBNAME).C(COLNAME).Find(filter)
}

func getSessionAccount(w http.ResponseWriter, r *http.Request)string{
	session,err := store.Get(r,SESSION_NAME)
	if err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return ""
	}

	accountObj := session.Values[ACCOUNT_SESSION_KEY]

	if accountObj == nil{
		httpError(w,"You are not logged in",http.StatusUnauthorized)
		return ""
	}

	accountName, ok := accountObj.(string)

	if !ok || accountName == ""{
		httpError(w,"Corrupt session",http.StatusInternalServerError)
		return ""
	}

	return accountName
}
//begin handlers

func sessionStatus(w http.ResponseWriter, r *http.Request){
	getSessionAccount(w,r)
}
func login(w http.ResponseWriter, r *http.Request){
	session,err := store.Get(r,SESSION_NAME)
	if err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}

	if session.Values[ACCOUNT_SESSION_KEY] != nil{
		httpError(w,"You are already logged in",http.StatusBadRequest)
		return
	}

	var loginReq struct{
		Username string
		Password string
	}

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&loginReq); err != nil{
		httpError(w,err.Error(),http.StatusBadRequest)
		return
	}

	if loginReq.Password != DUMMY_PASSWORD{
		httpError(w,"Bad password",http.StatusUnauthorized)
		return
	}
	
	count,err := getAlertsByAccount(loginReq.Username).Count()

	if err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}

	if count == 0{
		httpError(w,"This account does not exist",http.StatusUnauthorized)
		return
	}

	session.Values[ACCOUNT_SESSION_KEY] = loginReq.Username
	session.Save(r,w)
} 
func logoutSession(w http.ResponseWriter, r *http.Request){
	session,err := store.Get(r,SESSION_NAME)
	if err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}
	session.Values[ACCOUNT_SESSION_KEY] = nil
	session.Save(r,w)
}

func getAlerts(w http.ResponseWriter, r *http.Request){
	accountName := getSessionAccount(w,r)
	if accountName == ""{
		return
	}
	
	limit,err := strconv.Atoi(r.FormValue("limit"))
	if err != nil{
		limit = 3
	}

	offset,err := strconv.Atoi(r.FormValue("offset"))
	if err != nil{
		offset = 0
	}

	
	iter := getAlertsByAccount(accountName).Sort("-alert_data.ts_millis").Skip(offset).Limit(limit).Iter()

	var result []bson.M;

	if err := iter.All(&result); err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}

	enc := json.NewEncoder(w)

	if err := enc.Encode(result); err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}
}
func SyncHandler(w http.ResponseWriter,r *http.Request){
	if err := r.ParseMultipartForm(10*1000*1000); err != nil{
		httpError(w,err.Error(),http.StatusBadRequest)
		return
	}
	jsonData := r.FormValue("json_data")
	dec := json.NewDecoder(strings.NewReader(jsonData))

	var alertData map[string]interface{}

	if err := dec.Decode(&alertData); err != nil{
		httpError(w,err.Error(),http.StatusBadRequest)
		return
	}
	log.Printf("Alert Data: %+v",alertData)
	resourceName := r.FormValue("resource_name")
	accountName := r.FormValue("account_name")

	log.Printf("%s %s",resourceName,accountName)

	//yeah security
	if(resourceName != COLNAME){
		httpError(w,"invalid resouce_name",http.StatusBadRequest)
		return
	}

	//copy the picture file first
	file,fh,err := r.FormFile("file")
	
	if err != nil{
		httpError(w,err.Error(),http.StatusBadRequest)
		return
	}

	defer file.Close()

	filename := fmt.Sprintf("%d",rand.Uint32())+"_"+fh.Filename
	thumbname := "thumb."+filename

	baseUri := "pic_cache/"+accountName+"/"
	fileUri := baseUri+filename
	thumbUri := baseUri+thumbname

	fileDir := BASE_PIC_DIR+baseUri
	filePath := BASE_PIC_DIR+fileUri
	thumbPath := BASE_PIC_DIR+thumbUri

	if err := os.MkdirAll(fileDir,0775); err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}
		
	dst, err := os.Create(filePath)
	defer dst.Close()
	if err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return;
	}

	if _,err := io.Copy(dst,file); err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}
	
	log.Printf("finished receiving %s",filePath)

	//compress image
	cmd := exec.Command("ffmpeg","-y","-i",filePath,"-s","842x618",thumbPath)

	if err := FFExec(fileDir,cmd); err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}

	log.Printf("finished compressing image %s",thumbPath)
	//insert database row

	alertRow := bson.M{
		"account_name" : accountName,
		"file_path" : filePath,
		"file_uri" : fileUri,
		"thumb_path" : thumbPath,
		"thumb_uri" : thumbUri,
		"alert_data" : alertData,
		"sync_ts" : time.Now(),
	}

	resourceCol := dbsession.DB(DBNAME).C(resourceName)
	if err := resourceCol.Insert(&alertRow); err != nil{
		httpError(w,err.Error(),http.StatusInternalServerError)
		return
	}

	log.Println("inserted")
}
func FFExec(target string,cmd *exec.Cmd) error{
	
	stdout, err := cmd.StdoutPipe()

	if err != nil{
		return err
	}

	stderr, err := cmd.StderrPipe()

	if err != nil{
		return err
	}
	
	outfile, err := os.OpenFile(target+"/out.log",os.O_APPEND | os.O_WRONLY | os.O_CREATE,0700)

	if err != nil{
		return err
	}
	defer outfile.Close()
		
	errfile, err := os.OpenFile(target+"/err.log",os.O_APPEND | os.O_WRONLY | os.O_CREATE,0700)

	if err != nil{
		return err
	}
	defer errfile.Close()
	if _ , err := outfile.Write([]byte("\n\n"+time.Now().String()+"\n")); err != nil{
		return err
	}

	if _, err := errfile.Write([]byte("\n\n"+time.Now().String()+"\n")); err != nil{
		return err
	}

	go io.Copy(outfile,stdout)

	go io.Copy(errfile,stderr)

	// blocking for now
	if err := cmd.Run(); err != nil{
		return err
	}
	
	return nil
}
