package inspector

import (
	"log"
	"time"
	"io"
	"os"
	"os/exec"
	"bytes"
	"mime/multipart"

	"github.com/satori/go.uuid"

	"model"
)

func monitorResult(pythonScript string, duration, interval uint32){
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

func initMonitoringArray(pythonScript, title string){
	initAssertion := model.AssertionResult {
		PyScript:  		pythonScript,
		Title:          title,
		Timelog:        time.Now().Format(time.RFC3339),
		Times: 			0,
		Cleared:  	    false}
	log.Printf("init data : %v", initAssertion)
	broadcast <- initAssertion
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
	asserted.Timelog = time.Now().Format(time.RFC3339)
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

func saveCompanionFile(file multipart.File, uuid, filename string){
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
		// Grab the next Assertion from the broadcast channel
		assertionData := <-broadcast
		contain := monitoringData[assertionData.PyScript]
		contain = append(contain, assertionData)
		monitoringData[assertionData.PyScript] = contain
		for uuid := range clients {
            err := clients[uuid].Connection.WriteJSON(monitoringData)
            if err != nil {
                clients[uuid].Connection.Close()
                delete(clients, uuid)
            }
        }
	}
}