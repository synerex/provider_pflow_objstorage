package main

import (
	"encoding/json"
	"flag"
	"fmt"

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
)

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

// called for each agent data.
func supplyPFlowCallback(clt *sxutil.SXServiceClient, sp *api.Supply) {

	pc := &pflow.PFlow{}

	err := proto.Unmarshal(sp.Cdata.Entity, pc)
	if err == nil { // get PFlow
		tsd, _ := ptypes.Timestamp(pc.StartTime)

		// how to define Bucket:

		bucketName := "pflow"
		// we use IP address for sensor_id
		//		objectName := "area/year/month/date/hour/min"
		objectName := fmt.Sprintf("%s/%4d/%02d/%02d/%02d/%02d", pc.Area, tsd.Year(), tsd.Month(), tsd.Day(), tsd.Hour(), tsd.Minute())

		data, err := json.Marshal(pc)

		if err == nil {
			// log.Printf("Storing: %+v\n", data)
			// pfblocks に登録
			// 30秒程度ごとに pfblocks を監視
			// prevlen と長さが同じになったもの かつ 最低1分は経った block から objStore
			objStore(bucketName, objectName, string(data)+"\n")
		} else {
			log.Printf("Error!!: %+v\n", err)
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

	wg.Wait()

}
