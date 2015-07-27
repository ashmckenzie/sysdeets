package main

import (
  "runtime"
  "reflect"
  "log"
  "fmt"
  "time"
  "net/http"
  "github.com/shirou/gopsutil/disk"
  "github.com/shirou/gopsutil/mem"
  "github.com/shirou/gopsutil/cpu"
  "github.com/davecgh/go-spew/spew"
  "github.com/ant0ine/go-json-rest/rest"
  // "github.com/oleiade/reflections"
)

type Message struct {
  Body string
}

type Data struct {
	Memory Memory `json:"memory"`
	CPU    CPU    `json:"cpu"`
	Disk   Disk   `json:"disk"`
}

type Memory struct {
  Total string `json:"total"`
  Free  string `json:"free"`
}

type CPU struct {
  Count int32   `json:"count"`
  Idle  float64 `json:"idle"`
}

type Disk struct {
  MountPoint string `json:"mount_point"`
  Device     string `json:"device"`
  Total      uint64 `json:"total"`
  Free       uint64 `json:"free"`
}

var v *mem.VirtualMemoryStat
var ct []cpu.CPUTimesStat
var ci []cpu.CPUInfoStat
var disks []disk.DiskPartitionStat
var disksUsage *disk.DiskUsageStat

var err error

func updateMemory() {
  v, err = mem.VirtualMemory()
  log.Printf("updateMemory(): %v", spew.Sdump(v))
  if err != nil { log.Fatal(err) }
}

func updateCPUTimes() {
  ct, err = cpu.CPUTimes(false)
  log.Printf("updateCPUTimes(): %v", spew.Sdump(ct))
  if err != nil { log.Fatal(err) }
}

func updateCPUInfo() {
  ci, err = cpu.CPUInfo()
  log.Printf("updateCPUInfo(): %v", spew.Sdump(ci))
  if err != nil { log.Fatal(err) }
}

func updateDisk() {
  disks, err = disk.DiskPartitions(true)
  log.Printf("updateDisk(): %v", spew.Sdump(disks))
  if err != nil { log.Fatal(err) }

  disksUsage, err = disk.DiskUsage("/")
  log.Printf("updateDisk(): %v", spew.Sdump(disksUsage))
  if err != nil { log.Fatal(err) }
}

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
      Total: fmt.Sprintf("%d", v.Available),
      Free:  fmt.Sprintf("%d", v.Free),
    },
    CPU: CPU{
      Count: ci[0].Cores,
      Idle:  ct[0].Idle,
    },
    Disk: Disk{
      MountPoint: disks[0].Mountpoint,
      Device:     disks[0].Device,
      Total:      disksUsage.Total,
      Free:       disksUsage.Free,
    },
  }
  return d
}

func main() {
  log.Printf("Starting up..")

  updateMemory()
  updateCPUTimes()
  updateCPUInfo()
  updateDisk()

  api := rest.NewApi()
  api.Use(rest.DefaultDevStack...)

  go refreshData(updateMemory, 15)
  go refreshData(updateCPUTimes, 15)
  go refreshData(updateDisk, 60)

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
