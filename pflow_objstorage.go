package main

import (
	// "encoding/json"
	"flag"
	"fmt"
	"sort"
	"strings"
	"time"

	pflow "github.com/UCLabNU/proto_pflow"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	storage "github.com/synerex/proto_storage"
	api "github.com/synerex/synerex_api"
	pbase "github.com/synerex/synerex_proto"

	sxutil "github.com/synerex/synerex_sxutil"
	//sxutil "local.packages/synerex_sxutil"

	"log"
	"sync"
)

// datastore provider provides Datastore Service.

var (
	nodesrv         = flag.String("nodesrv", "127.0.0.1:9990", "Node ID Server")
	local           = flag.String("local", "", "Local Synerex Server")
	mu              sync.Mutex
	version         = "0.01"
	baseDir         = "store"
	dataDir         string
	pcMu            *sync.Mutex = nil
	pcLoop          *bool       = nil
	ssMu            *sync.Mutex = nil
	ssLoop          *bool       = nil
	sxServerAddress string
	currentNid      uint64                  = 0 // NotifyDemand message ID
	mbusID          uint64                  = 0 // storage MBus ID
	storageID       uint64                  = 0 // storageID
	pfClient        *sxutil.SXServiceClient = nil
	stClient        *sxutil.SXServiceClient = nil
	pfblocks        map[string]*PFlowBlock  = map[string]*PFlowBlock{}
	bucketName                              = flag.String("bucket", "centrair", "Bucket Name")
	holdPeriod                              = flag.Int64("holdPeriod", 720, "Flow Data Hold Time")
)

const layout = "2006-01-02T15:04:05.999999Z"

func init() {
}

func objStore(bc string, ob string, dt string) {

	log.Printf("Store %s, %s, %s", bc, ob, dt)
	//  we need to send data into mbusID.
	record := storage.Record{
		BucketName: bc,
		ObjectName: ob,
		Record:     []byte(dt),
		Option:     []byte("raw"),
	}
	out, err := proto.Marshal(&record)
	if err == nil {
		cont := &api.Content{Entity: out}
		smo := sxutil.SupplyOpts{
			Name:  "Record", // command
			Cdata: cont,
		}
		stClient.NotifySupply(&smo)
	}

}

// saveRecursive : save to objstorage recursive
func saveRecursive(client *sxutil.SXServiceClient) {
	// ch := make(chan error)
	for {
		time.Sleep(time.Second * time.Duration(60))
		currentTime := time.Now().Unix() + 9*3600
		log.Printf("\nCurrent: %d", currentTime)
		for name, pfblock := range pfblocks {
			if pfblock.BaseDate+*holdPeriod < currentTime {
				// data, err := json.Marshal(pfblock.PFlows)
				csvData := []string{}
				for _, pf := range pfblock.PFlows {
					st, _ := time.Parse(layout, ptypes.TimestampString(pf.Operation[0].Timestamp))
					et, _ := time.Parse(layout, ptypes.TimestampString(pf.Operation[1].Timestamp))
					csvData = append(csvData, fmt.Sprintf("%s,%s,%d,%d,%d,%d,%d", st.Format(layout), et.Format(layout), pf.Operation[0].Sid, pf.Operation[1].Sid, pf.Operation[0].Height, pf.Operation[1].Height, pf.Id))
				}

				// if err == nil {
				sort.Strings(csvData)
				objStore(*bucketName, name, strings.Join(csvData, "\n")+"\n")
				delete(pfblocks, name)
				// 	} else {
				// 		log.Printf("Error!!: %+v\n", err)
				// 	}
			}
		}
	}
}

// called for each agent data.
func supplyPFlowCallback(clt *sxutil.SXServiceClient, sp *api.Supply) {

	pc := &pflow.PFlow{}

	err := proto.Unmarshal(sp.Cdata.Entity, pc)
	if err == nil { // get PFlow
		tsd, _ := ptypes.Timestamp(pc.Operation[0].Timestamp)

		// how to define Bucket:

		// we use IP address for sensor_id
		//		objectName := "area/year/month/date/hour/min"
		objectName := fmt.Sprintf("%s/%s/%4d/%02d/%02d/%02d/%02d", "PFLOW", pc.Area, tsd.Year(), tsd.Month(), tsd.Day(), tsd.Hour(), tsd.Minute())

		if pfblock, exists := pfblocks[objectName]; exists {
			pfblock.PFlows = append(pfblock.PFlows, pc)
		} else {
			pfblocks[objectName] = &PFlowBlock{
				BaseDate: tsd.Unix(),
				PFlows:   []*pflow.PFlow{pc},
			}
		}
	}
}

func main() {
	flag.Parse()
	go sxutil.HandleSigInt()
	sxutil.RegisterDeferFunction(sxutil.UnRegisterNode)
	log.Printf("PFlow-ObjStorage(%s) built %s sha1 %s", sxutil.GitVer, sxutil.BuildTime, sxutil.Sha1Ver)

	channelTypes := []uint32{pbase.PEOPLE_FLOW_SVC, pbase.STORAGE_SERVICE}

	var rerr error
	sxServerAddress, rerr = sxutil.RegisterNode(*nodesrv, "PFlowObjStorage", channelTypes, nil)

	if rerr != nil {
		log.Fatal("Can't register node:", rerr)
	}
	if *local != "" { // quick hack for AWS local network
		sxServerAddress = *local
	}
	log.Printf("Connecting SynerexServer at [%s]", sxServerAddress)

	wg := sync.WaitGroup{} // for syncing other goroutines

	client := sxutil.GrpcConnectServer(sxServerAddress)

	if client == nil {
		log.Fatal("Can't connect Synerex Server")
	}

	stClient = sxutil.NewSXServiceClient(client, pbase.STORAGE_SERVICE, "{Client:PFObjStore}")
	pfClient = sxutil.NewSXServiceClient(client, pbase.PEOPLE_FLOW_SVC, "{Client:PflowStore}")

	log.Print("Subscribe PFlow Supply")
	pcMu, pcLoop = sxutil.SimpleSubscribeSupply(pfClient, supplyPFlowCallback)
	wg.Add(1)

	go saveRecursive(pfClient)

	wg.Wait()

}
