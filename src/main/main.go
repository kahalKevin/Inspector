package main

import (
	"log"
	"net/http"
	"time"
	"runtime"
	"strings"
	"strconv"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"bytes"
	"mime/multipart"
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"

	"model"
)

var isFinished = make(map[string]bool) // script in process
var monitoringData = make(map[string]model.AssertionResults)
var clients = make(map[string]model.User) // connected clients
var broadcast = make(chan model.AssertionResult)  // broadcast channel
var fileDir string
// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func init() {
	fileDir = "../file/"
    runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	// Create a simple file server
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/start", startAssertion)
	http.HandleFunc("/submit", submitAssertion)

	// Start listening for incoming assertion start and submision
	go handleAssertionSubmission()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()
	newUuid := uuid.NewV4()
	newUser := model.NewUser(newUuid.String(), ws, time.Now())
	// Register our new client
	clients[newUser.Id] = newUser
	initData := model.Message{
		TypeMsg:	"init",
		From   :	newUser.Id}

    log.Printf("A User Come")

	// Send init data to client, with some contact
	newUser.Connection.WriteJSON(initData)
}

func startAssertion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	usingCompanionFile := true

	//only accept 500kb file input
	if r.ContentLength > 500000 {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

    title := r.FormValue("title")
    duration := r.FormValue("duration")
    interval := r.FormValue("interval")
    file, header, err := r.FormFile("pythonscript")
    if err != nil {
    	w.WriteHeader(http.StatusPreconditionFailed)
        log.Printf("No file loaded")
        return
    }
    defer file.Close()
    fileExtension := strings.Split(header.Filename, ".")[1]

    file2, header2, err2 := r.FormFile("companionfile")
    if err2 != nil {
    	usingCompanionFile = false
        log.Printf("Not using companion file")
    }

    if(title=="" || duration=="" || interval=="" || fileExtension!="py"){
    	w.WriteHeader(http.StatusUnprocessableEntity)
    	return
    }

    if usingCompanionFile {
	    defer file2.Close()
	    fileExtension2 := strings.Split(header2.Filename, ".")[1]

	    if(fileExtension2!="txt" && fileExtension2!="csv"){
	    	w.WriteHeader(http.StatusUnprocessableEntity)
	    	return
	    }
    }

    var testDuration, testInterval uint32
    uintDuration, err := strconv.ParseUint(duration,10,32)
    uintInterval, err2 := strconv.ParseUint(interval,10,32)
	if  err==nil && err2==nil && uintDuration>=uintInterval && uintInterval<=10 && uintDuration<=120{
		testDuration = uint32(uintDuration)
		testInterval = uint32(uintInterval)
	}else{
		w.WriteHeader(http.StatusUnprocessableEntity)
    	return
	}

    scriptUuid := savePythonScript(file)
    if usingCompanionFile {
    	go saveCompanionFile(file2, scriptUuid, header2.Filename)
    }
    isFinished[scriptUuid] = false

    go initMonitoringArray(scriptUuid)
    go monitorResult(scriptUuid, testDuration, testInterval)

    log.Printf(scriptUuid)

	w.WriteHeader(http.StatusOK)
}

func monitorResult(pythonScript string, duration uint32, interval uint32){
	start := time.Now()
	for !isFinished[pythonScript]{
		elapsed := time.Since(start)
		elapsedMinutes := elapsed.Minutes()
		intElapsed := int(elapsedMinutes)
		if uint32(intElapsed) < duration{
			timer := time.NewTimer(time.Minute * time.Duration(interval))
			<-timer.C
			if isFinished[pythonScript] {
				go os.RemoveAll(fileDir+pythonScript+"/")
				return
			}
            cmd := exec.Command("python", fileDir+pythonScript+"/"+pythonScript+".py")
            stdout, err := cmd.Output()
            if(err!=nil){
            	log.Printf("Python script execution error: %v",err)
            }
            log.Printf(string(stdout))
		}else{
			isFinished[pythonScript]=true
		}
	}
	go os.RemoveAll(fileDir+pythonScript+"/")
}

func initMonitoringArray(pythonScript string){
	initAssertion := model.AssertionResult {
		PyScript:  		pythonScript,
		Times: 			0,
		Cleared:  	    false}
	log.Printf("init data : %v", initAssertion)
	broadcast <- initAssertion
}

func submitAssertion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//limit incoming result to 500kb
	body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 500000))

	var asserted model.AssertionResult
	json.Unmarshal(body, &asserted)

	if _, exist := isFinished[asserted.PyScript]; exist{		
		if reviewAssertionDone(asserted.Asserted){
			isFinished[asserted.PyScript]=true
			log.Printf("Script %s has finished", asserted.PyScript)
		}
		go insertMonitoringDataToArray(asserted, isFinished[asserted.PyScript])
		w.WriteHeader(http.StatusOK)
	}else{
		w.WriteHeader(http.StatusNotFound)
	}
}

func reviewAssertionDone(asserted model.AssertionEntities) bool{
	finished := true
	for _, content := range asserted {
		if !content.Success{
			finished = false
			break
		}
	}
	return finished
}

func insertMonitoringDataToArray(asserted model.AssertionResult, succeed bool){
	asserted.Times = len(monitoringData[asserted.PyScript])
	asserted.Cleared = succeed
	log.Printf("insert updated data : %v", asserted)
	broadcast <- asserted
}

func savePythonScript(file multipart.File) string{
	var Buf bytes.Buffer
    io.Copy(&Buf, file)
    contents := string(Buf.Bytes())
    Buf.Reset()
    newUuid := uuid.NewV4().String()

    os.Mkdir(fileDir+newUuid, 0777)
    f, _ := os.Create(fileDir+newUuid+"/"+newUuid+".py")
    defer f.Close()

    _, err := f.WriteString(contents)
    if(err!=nil){
    	log.Printf("%v",err)
    }
    f.Sync()

    return newUuid
}

func saveCompanionFile(file multipart.File, uuid string, filename string){
	var Buf bytes.Buffer
    io.Copy(&Buf, file)
    contents := string(Buf.Bytes())
    Buf.Reset()

    f, _ := os.Create(fileDir+uuid+"/"+filename)
    defer f.Close()

    _, err := f.WriteString(contents)
    if(err!=nil){
    	log.Printf("%v",err)
    }
    f.Sync()
}

func handleAssertionSubmission() {
	for {
		// Grab the next message from the broadcast channel
		assertionData := <-broadcast
		contain := monitoringData[assertionData.PyScript]
		contain = append(contain, assertionData)
		monitoringData[assertionData.PyScript] = contain
		for uuid := range clients {
            err := clients[uuid].Connection.WriteJSON(monitoringData)
            if err != nil {
                log.Printf("error: %v", err)
                clients[uuid].Connection.Close()
                delete(clients, uuid)
            }
        }
	}
}