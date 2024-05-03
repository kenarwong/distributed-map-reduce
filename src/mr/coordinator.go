package mr

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"
)

// const INPUT_SLICE_SIZE_MB = 64
// const INPUT_SLICE_SIZE_BYTES = 64 * 1024

const COORDINATOR_INIT_PHASE = 0
const COORDINATOR_MAP_PHASE = 1
const COORDINATOR_REDUCE_PHASE = 2
const COORDINATOR_COMPLETE_PHASE = 3

const WORKER_STATE_IDLE = 0
const WORKER_STATE_ACTIVE = 1

type Coordinator struct {
	// Your definitions here.
	phase      int
	nReduce    int
	workers    map[int]TaskWorker
	toDo       map[int]Task
	inProgress map[int]Task
	completed  map[int]Task
}

type TaskWorker struct {
	state    int
	taskType int
	task     Task
}

// func (c *Coordinator) () {
// }

// Your code here -- RPC handlers for the worker to call.

// an example RPC handler.
//
// the RPC argument and reply types are defined in rpc.go.
func (c *Coordinator) Example(args *ExampleArgs, reply *ExampleReply) error {
	reply.Y = args.X + 1
	return nil
}

func (c *Coordinator) Register(args *interface{}, reply *RegisterReply) error {
	// Phase restrictions

	// Register worker ID
	workerId := len(c.workers)
	c.workers[workerId] = TaskWorker{
		state:    WORKER_STATE_IDLE,
		taskType: NO_TASK_TYPE,
	}

	// Reply
	// fmt.Println("worker id:", workerId)
	reply.WorkerId = workerId
	// fmt.Println("reply worker id:", reply.WorkerId)
	return nil
}

func (c *Coordinator) GetTask(args *TaskArgs, reply *TaskReply) error {
	// Phase restrictions
	switch c.phase {
	case COORDINATOR_MAP_PHASE:
	case COORDINATOR_REDUCE_PHASE:
	default:
		err := fmt.Errorf("invalid phase %v", c.phase)
		return err
	}

	switch c.phase {
	case MAP_TASK_TYPE:
		// Get task
		i := len(c.toDo) - 1

		if i < 0 {
			log.Printf("No more tasks.")
			reply.WorkerId = args.WorkerId
			reply.TaskType = NO_TASK_TYPE
			return nil
		}

		t, ok := c.toDo[i]
		if !ok {
			// log.Fatalf("error %v", err)
			log.Printf("GetTask (Coordinator): Missing task at index %v", i)
			err := fmt.Errorf("Task tracking error")
			return err
		}
		delete(c.toDo, i)

		fmt.Printf("task %v\n", t.(MapTask).id)

		// Reply to worker
		reply.WorkerId = args.WorkerId
		reply.TaskType = MAP_TASK_TYPE
		reply.MapTaskId = t.(MapTask).id
		reply.Filename = t.(MapTask).inputSlice.filename
		reply.ReportInterval = 5

		// Move task to inProgress
		//t.startTime = time.Now()
		c.inProgress[i] = t

		// Update worker state
		c.workers[args.WorkerId] = TaskWorker{
			state:    WORKER_STATE_ACTIVE,
			taskType: MAP_TASK_TYPE,
			task:     t,
		}

	case REDUCE_TASK_TYPE:
	}

	return nil
}

func (c *Coordinator) TaskStatus(args *TaskStatusArgs, reply *TaskStatusReply) error {
	// Phase restrictions
	switch c.phase {
	case COORDINATOR_MAP_PHASE:
	case COORDINATOR_REDUCE_PHASE:
	default:
		err := fmt.Errorf("invalid phase %v", c.phase)
		return err
	}

	switch args.TaskType {
	case MAP_TASK_TYPE:
		// Task report
		fmt.Printf("Report: WorkerId %v, TaskType %v, MapTaskId %v \n", args.WorkerId, args.TaskType, args.MapTaskId)
		fmt.Printf("Report: Progress %v, Complete %v \n", args.Progress, args.Complete)

		// Complete
		if args.Complete {
			// Move task from inProgress to Complete
			//t.lastUpdatedTime = time.Now()

			t, ok := c.inProgress[args.MapTaskId]
			if !ok {
				log.Printf("GetTask (Coordinator): Missing task at index %v", args.MapTaskId)
				err := fmt.Errorf("Task tracking error")
				return err
			}

			// Remove from inProgress
			delete(c.inProgress, args.MapTaskId)
			fmt.Println("In Progress:")
			fmt.Println(c.inProgress)

			// Add to completed
			c.completed[args.MapTaskId] = t
			fmt.Println("Completed:")
			fmt.Println(c.completed)

			// Update worker state
			c.workers[args.WorkerId] = TaskWorker{
				state:    WORKER_STATE_IDLE,
				taskType: NO_TASK_TYPE,
				task:     nil,
			}
		}

		// In Progress
		// Update task progress

		// Check if no progress made
		// No progress made
		// Check against timeout
		// Move task back to toDo

	case REDUCE_TASK_TYPE:
	default:
		log.Printf("TaskStatus: Unknown TaskType %v\n", args.TaskType)
	}

	return nil
}

func (c *Coordinator) InitTasks(files []string) error {
	// Calculate input slices
	fmt.Println("MakeCoordinator (Coordinator): Calculating files...")
	fmt.Printf("Number of files: %d\n", len(files))
	fmt.Println("MakeCoordinator (Coordinator): Creating tasks...")
	for i, filename := range files {
		//fmt.Printf("id: %d, file %v\n", i, filename)

		t := MapTask{
			id:         i,
			inputSlice: InputSlice{filename: filename},
		}
		c.toDo[i] = t

		//	fmt.Println(filename)
		//	fi, err := os.Stat(filename)
		//	if err != nil {
		//		// Could not obtain stat, handle error
		//		log.Fatalf("error stat %v", filename)
		//	}
		//	fmt.Printf("The file is %d bytes long", fi.Size())

		//	//file, err := os.Open(filename)
		//	//if err != nil {
		//	//	log.Fatalf("cannot open %v", filename)
		//	//}

		//	//bArr := make([]byte, sliceSize)
		//	//for {
		//	//	n, err := file.Seek(head, tail)

		//	//	slice := InputSlice{filename: filename}}

		//	//	if err == io.EOF {
		//	//		break
		//	//	}
		//	//}

	}
	//fmt.Println("ToDo:", c.toDo)

	return nil
}

func (c *Coordinator) StatusCheck() error {
	// Phase restrictions
	switch c.phase {
	case COORDINATOR_MAP_PHASE:
		fmt.Println("StatusCheck (Coordinator): Map phase")

		// Check ToDo and InProgress
		if len(c.toDo) == 0 && len(c.inProgress) == 0 {
			// Worker status report
			for w, wk := range c.workers {
				if wk.state != WORKER_STATE_IDLE {
					fmt.Printf("Worker Id %v not idle", w)
				}
			}

			// Change to reduce phase
			fmt.Println("StatusCheck (Coordinator): Change to reduce phase")
			c.phase = COORDINATOR_REDUCE_PHASE
		}
	case COORDINATOR_REDUCE_PHASE:
		fmt.Println("StatusCheck (Coordinator): Reduce phase")

		// Check ToDo and InProgress
		// Check reduce slice completion status

		// Change to complete phase when reduce phase complete
		fmt.Println("StatusCheck (Coordinator): Change to complete phase")
		c.phase = COORDINATOR_COMPLETE_PHASE
	default:
		err := fmt.Errorf("invalid phase %v", c.phase)
		return err
	}

	return nil
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	ret := false // Return true for mrcoordinator.go to exit

	fmt.Println("Looping...")
	time.Sleep(5)

	// Heartbeats from workers

	// Check status
	c.StatusCheck()

	// Set ret to true when reduce phase complete
	if c.phase == COORDINATOR_COMPLETE_PHASE {
		ret = true
	}
	return ret
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{
		phase:      COORDINATOR_INIT_PHASE,
		nReduce:    nReduce,
		workers:    make(map[int]TaskWorker),
		toDo:       make(map[int]Task),
		inProgress: make(map[int]Task),
		completed:  make(map[int]Task),
	}

	// Initialize tasks
	c.InitTasks(files)

	// Start RPC server
	c.phase = COORDINATOR_MAP_PHASE
	go c.server()
	return &c
}
