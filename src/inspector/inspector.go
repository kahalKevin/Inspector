package inspector

import (
	"log"
	"net/http"
	"time"
	"strings"
	"strconv"
	"io"
	"io/ioutil"
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"

	"model"
	"mailer"
)

var isFinished = make(map[string]bool) // script in process
var monitoringData = make(map[string]model.AssertionResults)
var clients = make(map[string]model.User) // connected clients
var broadcast = make(chan model.AssertionResult)  // broadcast channel
var fileDir, webPort string
var Mailer   mailer.Mailer
// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Start(dir, port, host, sender, public string, smtpPort int, recpt []string, subject, body map[string]string) {
	fileDir = dir
	webPort = port
	// Create a simple file server
	fs := http.FileServer(http.Dir(public))
	http.Handle("/", fs)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/start", startAssertion)
	http.HandleFunc("/submit", submitAssertion)
	http.HandleFunc("/stop", stopAssertion)
	http.HandleFunc("/getdata",getMonitoringData)

	// Start listening for incoming assertion start and submision
	go handleAssertionSubmission()

	Mailer = mailer.New(
		host,
		smtpPort,
		sender,
		recpt,
	).SetMessages(subject, body)

	log.Println("http server started on", webPort)
	err := http.ListenAndServe(webPort, nil)
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
	// Connection will be closed later if Inspector cant send data anymore
	// defer ws.Close()

	newUuid := uuid.NewV4()
	newUser := model.NewUser(newUuid.String(), ws, time.Now())
	// Register our new client
	clients[newUser.Id] = newUser
    log.Printf("A User Come")

	// Send current monitoring data to client
	newUser.Connection.WriteJSON(monitoringData)
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

    go initMonitoringArray(scriptUuid, title)
    go monitorResult(scriptUuid, testDuration, testInterval)

    log.Printf(scriptUuid)

	w.WriteHeader(http.StatusOK)
}

func submitAssertion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	//limit incoming result to 500kb
	body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 500000))

	var asserted model.AssertionResult
	json.Unmarshal(body, &asserted)
	if strings.Contains(asserted.PyScript, fileDir){
		splittedPath := strings.Split(asserted.PyScript, "/")
		asserted.PyScript = splittedPath[len(splittedPath)-1]
	}
	if _, exist := isFinished[asserted.PyScript]; exist{
		if reviewAssertionDone(asserted.Asserted){
			isFinished[asserted.PyScript]=true
			log.Printf("Script %s has finished", asserted.PyScript)
		}
		go insertMonitoringDataToArray(asserted, isFinished[asserted.PyScript])
		w.WriteHeader(http.StatusOK)
	}else{
		log.Printf(asserted.PyScript)
		w.WriteHeader(http.StatusNotFound)
	}
}

func stopAssertion(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    scriptUuid := r.URL.Query().Get("script")
    if _, exist := isFinished[scriptUuid]; exist{
		isFinished[scriptUuid]=true
		log.Printf("Script %s has stopped by command", scriptUuid)
		broadcast <- model.AssertionResult{
                      PyScript:     scriptUuid,
                      Timelog:      time.Now().Format(time.RFC3339),
                      Times:        -1,
                      Cleared:      false}
		w.WriteHeader(http.StatusOK)
	}else{
		w.WriteHeader(http.StatusNotFound)
	}
}

func getMonitoringData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(monitoringData)
}