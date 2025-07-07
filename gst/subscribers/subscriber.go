package subscribers

import (
        "fmt"
        "mm/queue"
        "mm/services/gstmmcontrols"
        "mm/services/masterindiacontrols"
        "mm/services/gstchallandata"
)

type workRequest struct {
        Data       map[string]string
        RcptHandle *string
}

type worker struct {
        id          int
        work        chan workRequest
        workerQueue chan chan workRequest
}

var workerQueue chan chan workRequest
var workQueue chan workRequest

func callService(wr workRequest) {

        switch wr.Data["publisher"] {
        case "dhl":
                //fmt.Println("dhl", wr.Data["jsonDataStr"])
                gstmmcontrols.SubcriberHandler(wr.Data["jsonDataStr"])
        case "centralizedAPI":
                //fmt.Println("centralizedAPI", wr.Data["jsonDataStr"])
                masterindiacontrols.SubcriberHandler(wr.Data["jsonDataStr"])
        case "instantact":
                gstchallandata.SubcriberHandler(wr.Data["jsonDataStr"])
        }
}

func newWorker(id int, workerQueue chan chan workRequest) worker {

        return worker{
                id:          id,
                work:        make(chan workRequest),
                workerQueue: workerQueue,
        }
}

func (w *worker) start() {

        go func() {

                for {
                        w.workerQueue <- w.work

                        select {
                        case work := <-w.work:
                                fmt.Println(w.id, " got wr and doing ")
                                callService(work)
                                //fmt.Println(w.id, " deleting from queue", work.RcptHandle)
                                err := queue.Delete(work.RcptHandle)
                                fmt.Println("deletion status: ", err)
                        }

                }

        }()
}

//StartDispatcher ...
func StartDispatcher(nWorkers int, workQueueSize int) {

        workerQueue = make(chan chan workRequest, nWorkers)

        for i := 0; i < nWorkers; i++ {
                fmt.Println("Starting worker", i+1)
                worker := newWorker(i+1, workerQueue)
                worker.start()
        }

        workQueue := make(chan workRequest, workQueueSize)

        go func() {
                for {
                        var err error = nil
                        var rHandle *string
                        deqData, rHandle, _, err := queue.Receive()

                        if err != nil {
                                continue
                        }

                        workRequest := workRequest{
                                Data:       deqData,
                                RcptHandle: rHandle,
                        }

                        workQueue <- workRequest
                        //fmt.Println("Putting wr from sqs to work_queue")

                        select {

                        case work := <-workQueue:
                                //fmt.Println("got wr from work queue ")

                                go func() {
                                        worker := <-workerQueue
                                        //fmt.Println("Dispatching work request ")
                                        worker <- work
                                }()
                        }
                }
        }()
}
