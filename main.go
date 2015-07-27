package main

import (
  "runtime"
  "reflect"
  "log"
  "fmt"
  "time"
  "net/http"
  "github.com/shirou/gopsutil/mem"
  // "github.com/shirou/gopsutil/cpu"
  "github.com/davecgh/go-spew/spew"
  "github.com/ant0ine/go-json-rest/rest"
  // "github.com/oleiade/reflections"
)

type Message struct {
  Body string
}

type Data struct {
	Memory Memory `json:"memory"`
	CPU CPU `json:"cpu"`
}

type Memory struct {
  Total string `json:"total"`
  Free  string `json:"free"`
}

type CPU struct {
  Count  int `json:"count"`
}

var v *mem.VirtualMemoryStat
// var c []cpu.CPUTimesStat
var err error

func updateMemory() {
  v, err = mem.VirtualMemory()
  spew.Dump(v)
  if err != nil { log.Fatal(err) }
}

// func updateCPU() {
//   c, err := cpu.CPUTimes(true)
//   spew.Dump(c)
//   spew.Dump(len(c))
//   if err != nil { log.Fatal(err) }
// }

func nameOf(f interface{}) string {
	v := reflect.ValueOf(f)
	if v.Kind() == reflect.Func {
		if rf := runtime.FuncForPC(v.Pointer()); rf != nil {
			return rf.Name()
		}
	}
	return v.String()
}

func refreshData(what func(), x time.Duration) {
  for _ = range time.Tick(time.Second * x) {
    log.Printf("Running %v()..", nameOf(what))
    what()
  }
}

func data() Data {
  d := Data{
    Memory: Memory{
      Total: fmt.Sprintf("%d", v.Total),
      Free: fmt.Sprintf("%d", v.Free),
    },
    CPU: CPU{
      Count: 2,
    },
  }
  return d
}

func main() {
  log.Printf("Starting up..")

  updateMemory()
  // updateCPU()

  api := rest.NewApi()
  api.Use(rest.DefaultDevStack...)

  go refreshData(updateMemory, 15)
  // go refreshData(updateCPU, 15)

  router, err := rest.MakeRouter(

    rest.Get("/", func(w rest.ResponseWriter, req *rest.Request) {
      d := data()
      w.WriteJson(d)
    }),

    // rest.Get("/#key", func(w rest.ResponseWriter, req *rest.Request) {
    //   data := data()
    //   extractedData, err := reflections.GetField(data, req.PathParam("key"))
    //
    //   if err != nil {
    //     rest.Error(w, err.Error(), http.StatusInternalServerError)
    //     return
    //   }
    //
    //   w.WriteJson(extractedData)
    // }),

  )

  if err != nil { log.Fatal(err) }

  api.SetApp(router)
  log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}
