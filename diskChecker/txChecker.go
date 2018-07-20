package main

import (
	"strings"
	"os"
	"bufio"
	"io"
	"fmt"
	"strconv"
	"flag"
)

//transaction struct.
type Tx struct {
	txStart string			//transaction start time.
	txEnd string            //transaction end time.
	txPeriod string			//transaction period.
	txHash string			//transaction hash.
	txSignPeriod string  	//This field is nil in SendRawTransaction.
	txerr string			//transaction error.

	addTxStart string		//addTx() start at.
	addTxEnd string			//addTx() end at.
	addTxPeriod string      //addTx() period.
	addTxHash string		//transaction hash in addTx.
	addTxPending string     //The count of transactions in pending queue.
	addTxLocal string		//Whether this transaction is local or non-local.
	addTxSender string      //The sender of the transaction.

	submitTxStart string	//submitTransactions() start time.
	submitTxEnd string      //submitTransactions() end time.
	submitTxPeriod string   //submitTransactions() period.
	submitTxHash string     //hash of the ransaction in submitTransactions.

	sendDone bool			//sendTransactions analysis is done.
	addDone bool			//addTx analysis is done.
	submitDone bool			//submitTransaction analysis is done.
}

//block transaction.
type Block struct {
	number string			//block number
	timestamp string        //block timestamp
	size string             //block size
	txCnt string            //block transaction count
	miner string            //miner
	start string			//time for commitNewWork() starting.
	end string              //time for commitNewWork() ending.
	period string           //time for commitNewWork() period.
	readPending string      //time for reading pending queue.
	sort string             //time for sorting by price and nonce.
	commitTxs string        //time for committing transactions.
	finalize string			//time for finalizing block.
	sealStart string		//time for seal starting.
	sealPeriod string		//time for seal period.
}

//blockInfo list.
var blockInfo = make(map[string] *Block)

var logFolder string = "/home/guoxu/disktest/"
var logFileList = [4]string{
	logFolder + "50001FilterResult.log",
	logFolder + "50002FilterResult.log",
	logFolder + "50003FilterResult.log",
	logFolder + "50004FilterResult.log",
}

const nodeCount = 4
var nodeId = [nodeCount]string {
	"0x4B9ED090E635cCe279Fe5Bd7948F93edEbb9E634",
	"0x0f0b004B1FA88Eb885091df840316a56fCb80dF4",
	"0x9d14e190712008129a440473C010B3FD7599153a",
	"0xBc6b6FC74A9379C75bBD8Cf3B6410a3AfC229570",
}

const (
	OptionTypeBlock string = "block"
	optionTypeTransaction string = "transaction"
	resultWriterBuffer int = 1
	isAddTxFlag = "addTx start="
	isSubmitTxFlag = "submitTransaction start="
	isSendTxFlag = "SendTransaction start="
	startFlag = "="
	endFlag = " "
	finalFlag = "\n"
)

const (
	IsAddTx = iota
	IsSubmitTx
	IsSendTx
	empty
)

func IsTxCheckerLine(l string) int {
	if(strings.Contains(l, isSendTxFlag)) {
		return IsSendTx
	} else if(strings.Contains(l,isAddTxFlag)) {
		return IsAddTx
	} else if(strings.Contains(l,isSubmitTxFlag)) {
		return IsSubmitTx
	}
	return empty
}

const maxLoop = 100
func parseLine(line string, pre string, post string) []string {
	items := make([]string,maxLoop)
	cur := line

	for  i:=0;i < maxLoop;i++  {
		if(!strings.Contains(cur,pre)) {
			break
		}

		begin := strings.Index(cur, pre)
		cur = cur[begin+1 : ]

		//deal with the final item.
		var end int
		if(strings.Contains(cur,finalFlag) && !strings.Contains(cur, pre)) {
			end = strings.Index(cur, finalFlag)
		} else {
			end = strings.Index(cur, post)
		}
		items[i] = cur[0 : end]
	}
	return items
}

func dealWithAddTx(line string,tx *Tx) {
	items := parseLine(line, startFlag, endFlag)

	tx.addTxStart = items[0]  //start.
	tx.addTxEnd = items[1]  //end.
	tx.addTxPeriod = items[2]  //period
	tx.addTxHash = items[3]  //hash.
	tx.addTxPending = items[4]  //size.
	tx.addTxLocal = items[5]  //local.
	tx.addTxSender = items[6]  //sender.
	tx.addDone = true
}

func dealWithSendTx(line string,tx *Tx)  {
	items := parseLine(line, startFlag, endFlag)

	tx.txStart = items[0]  //start.
	tx.txEnd = items[1]  //end.
	tx.txPeriod = items[2]  //end.
	tx.txHash = items[3]  //end.
	tx.txSignPeriod = items[4]  //end.
	tx.txerr = items[4] //end.
	tx.sendDone = true
}

func dealWithSubmitTx(line string,tx *Tx) {
	items := parseLine(line, startFlag, endFlag)

	tx.submitTxStart = items[0]  //start.
	tx.submitTxEnd = items[1]  //end.
	tx.submitTxPeriod = items[2]  //end.
	tx.submitTxHash = items[3]
	tx.submitDone = true
}

func isCompleted(tx *Tx) bool {
	if(tx.addDone && tx.sendDone && tx.submitDone) {
		if((strings.Compare(tx.submitTxHash,tx.txHash) == 0) && (strings.Compare(tx.submitTxHash,tx.addTxHash) == 0) && (strings.Compare(tx.txHash,tx.addTxHash) == 0)) {
			return true
		} else {
			return false
		}
	} else {
		fmt.Println("Tx is not completed, Do not write it to file.")
		return false
	}
}

func ParseBlockInfoLine(line string, from int,to int)  {
	line = strings.Replace(line,"\n","",-1)
	items := strings.Split(line, ",")
	number,err := strconv.Atoi(items[0])
	if(err != nil) {
		panic(err)
		return
	}

	if(number >= from && number <= to) {
		//store block to map.
		block := Block{}
		block.number = items[0]
		block.timestamp = items[1]
		block.size = items[2]
		block.txCnt = items[3]
		block.miner = items[4]

		blockInfo[items[0]] = &block
	}
}

func ParseBlockFile(file string,fromBlockNum int, toBlockNum int) {
	bInfoFile, err := os.Open(file)
	if(err != nil) {
		panic(err)
	}
	defer bInfoFile.Close()

	rd := bufio.NewReader(bInfoFile)
	for {
		line, err := rd.ReadString('\n')
		if(err != nil || io.EOF == err) {
			fmt.Println(err)
			break
		}

		//deal with line now.
		ParseBlockInfoLine(line,fromBlockNum,toBlockNum)
	}
	fmt.Printf("Total block count: %d\n", len(blockInfo))
}

func ParaseLogFileForTxs( file string, resultCh chan<- string)  {
	f, err := os.Open(file)
	if(err != nil) {
		panic(err)
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	tx := Tx{}
	indexNum := 1
	for {
		line,err := rd.ReadString('\n')
		if(err != nil || io.EOF == err) {
			fmt.Println(err)
			break
		}

		t := IsTxCheckerLine(line)
		switch t {
		case IsAddTx:
			dealWithAddTx(line,&tx)
		case IsSendTx:
			dealWithSendTx(line,&tx)
			if(isCompleted(&tx)) {
				str := strconv.Itoa(indexNum) 	+ "," +
					   tx.txHash 				+ "," +
					   tx.txStart 				+ "," +
					   tx.txEnd 				+ "," +
					   tx.txPeriod 				+ "," +
					   tx.txSignPeriod 			+ "," +
					   tx.submitTxStart 		+ "," +
					   tx.submitTxEnd 			+ "," +
					   tx.submitTxPeriod 		+ "," +
					   tx.addTxStart 			+ "," +
					   tx.addTxEnd 				+ "," +
					   tx.addTxPeriod 			+ "," +
					   tx.addTxPending 			+ "," +
					   tx.addTxLocal 			+ "," +
					   tx.addTxSender 			+
					   "\n"
				resultCh <- str
				tx = Tx{}
				indexNum++
			}
		case IsSubmitTx:
			dealWithSubmitTx(line,&tx)
		}
	}
	fmt.Println("Done.")
}

func DealwithAim1(line string) {
	items := parseLine(line, startFlag, endFlag)

	bNumber := items[0]
	_,exist := blockInfo[bNumber]
	if(!exist) {
		blockInfo[bNumber] = &Block{}
	}
	blockInfo[bNumber].number = bNumber  		//number.
	blockInfo[bNumber].start = items[1] 		//start
	blockInfo[bNumber].end = items[2] 			//end
	blockInfo[bNumber].period = items[3] 		//period
	blockInfo[bNumber].readPending = items[4] 	//pending.
	blockInfo[bNumber].sort = items[5] 			//sort.
	blockInfo[bNumber].txCnt = items[6] 		//commit txs.
	blockInfo[bNumber].finalize = items[7] 		//finalize.
}

func DealwithAim2(line string)  {
	items := parseLine(line,startFlag,endFlag)

	number := items[1]
	blockInfo[number].sealStart = items[0] //seal start.
}

func DealwithAim3(line string) string {
	items := parseLine(line,startFlag,endFlag)

	number := items[0]
	blockInfo[number].sealPeriod = items[2]
	blockInfo[number].miner = items[3]
	return number
}

func ParseLogFileForBlocks(targetFile string, resultCh chan<- string, fromBlock string, toBlock string)  {
	//Open log files and make file reader for per log.
	//var rds [nodeCount] *bufio.Reader
	//var fileHandles [nodeCount] *os.File

	/*
	for i:=0; i<nodeCount ;i++  {
		f,err := os.Open(logFileList[i])
		if(err != nil) {
			panic(err)
		}
		fileHandles[i] = f
		rds[i] = bufio.NewReader(fileHandles[i])
		defer fileHandles[i].Close()
	}
	*/

	//var rd *bufio.Reader

	//get the node which make the current block.
	start,_ := strconv.Atoi(fromBlock)
	end,_ := strconv.Atoi(toBlock)

	for i:=start;i<end;i++ {
		number := strconv.Itoa(i)
		block := blockInfo[number]
		miner := block.miner

		/*
		if(strings.Compare(miner, nodeId[0]) == 0) {
			fmt.Println("searching log 0")
			rd = rds[0]
		} else if(strings.Compare(miner, nodeId[1]) == 0) {
			fmt.Println("searching log 1")
			rd = rds[1]
		} else if(strings.Compare(miner,nodeId[2]) == 0) {
			fmt.Println("searching log 2")
			rd = rds[2]
		} else if(strings.Compare(miner, nodeId[3]) == 0) {
			fmt.Println("searching log 3")
			rd = rds[3]
		} else {
			fmt.Printf("Can not map miner address %s to node.\n" , miner)
			break
		}
		*/
		var fileName string
		if(strings.Compare(miner, nodeId[0]) == 0) {
			fmt.Println("searching log 1 for block " + number)
			fileName = logFileList[0]
		} else if(strings.Compare(miner, nodeId[1]) == 0) {
			fmt.Println("searching log 2 for block " + number)
			fileName = logFileList[1]
		} else if(strings.Compare(miner,nodeId[2]) == 0) {
			fmt.Println("searching log 3 for block " + number)
			fileName = logFileList[2]
		} else if(strings.Compare(miner, nodeId[3]) == 0) {
			fmt.Println("searching log 4 for block " + number)
			fileName = logFileList[3]
		} else {
			fmt.Printf("Can not map miner address %s to node.\n" , miner)
			break
		}

		f,err := os.Open(fileName)
		if(err != nil) {
			panic(err)
		}
		rd := bufio.NewReader(f)


		aimStr1 := "commitNewWork blockNum=" + number
		aimStr2_1 := "Seal start at="
		aimStr2_2 := "Num=" + number
		aimStr3 := "Seal blockNum=" + number
		aim1CompletedFlag := false
		aim2COmpletedFlag := false
		aim3CompletedFlag := false

		//Search the log of the certain node and find the comment line.
		for {
			line,err := rd.ReadString('\n')
			if(err != nil || io.EOF == err) {
				fmt.Println(err)
				break
			}

			if strings.Contains(line,aimStr1) {
				//dealwith aim1.
				DealwithAim1(line)
				aim1CompletedFlag = true
			} else if(strings.Contains(line,aimStr2_1) && strings.Contains(line,aimStr2_2)) {
				//dealwith aim2.
				DealwithAim2(line)
				aim2COmpletedFlag = true
			} else if(strings.Contains(line,aimStr3)) {
				//dealwith aim3.
				num := DealwithAim3(line)
				aim3CompletedFlag = true
				if(aim1CompletedFlag && aim2COmpletedFlag && aim3CompletedFlag) {
					//write to file.
					str := blockInfo[num].number 		+ "," +
						   blockInfo[num].timestamp 	+ "," +
						   blockInfo[num].miner 		+ "," +
						   blockInfo[num].size 			+ "," +
						   blockInfo[num].txCnt 		+ "," +
						   blockInfo[num].start 		+ "," +
						   blockInfo[num].end 			+ "," +
						   blockInfo[num].period 		+ "," +
						   blockInfo[num].readPending 	+ "," +
						   blockInfo[num].sort 			+ "," +
						   blockInfo[num].commitTxs 	+ "," +
						   blockInfo[num].finalize 		+ "," +
						   blockInfo[num].sealStart 	+ "," +
						   blockInfo[num].sealPeriod 	+
						   "\n"

					resultCh <- str
				}
				//delete block infomation from the map to save memory.
				//delete(blockInfo, k)	//delete item from the map.
				//fmt.Printf("delete block info from map %s\n", k)
				aim3CompletedFlag = false
				aim2COmpletedFlag = false
				aim1CompletedFlag = false
				break
			}
		}

		f.Close()
	}

}

func ResultWriter(file string, resultCh <-chan string) {
	//open block target file.
	target,err := os.OpenFile(file, os.O_CREATE|os.O_RDWR|os.O_APPEND ,0666)
	if(err != nil) {
		panic(err)
	}
	defer target.Close()
	wr := bufio.NewWriter(target)

	for {
		result := <-resultCh
		wr.WriteString(result)
		wr.Flush()
	}
}

func main() {
	//command parameters.

	//Test for block.
	optType := flag.String("type", "block","option 'type': block or transaction")
	optFromBlock := flag.String("from", "5500", "get block info from 'from' block number")
	optToBlock := flag.String("to","6900","get block info to 'to' block number")
	optBlockInfoFile := flag.String("blockInfoFile","/home/guoxu/disktest/blockInfo.txt","get block info from 'blockInfoFile'")
	optTranFile := flag.String("tranFile","/home/guoxu/disktest/50001FilterResult.log","get transaction information from tranFile")
	optTargetFile :=flag.String("targetFile","/home/guoxu/disktest/resultBlocks.txt","store analysis result to targetFile")

	//Test for transaction.
	/*
	optType := flag.String("type", "transaction","option 'type': block or transaction")
	optFromBlock := flag.String("from", "", "get block info from 'from' block number")
	optToBlock := flag.String("to","","get block info to 'to' block number")
	optBlockInfoFile := flag.String("blockInfoFile","","get block info from 'blockInfoFile'")
	optTranFile := flag.String("tranFile","/mnt/hgfs/vmware_share/gwan4nodesLog/gx1.log","get transaction information from tranFile")
	optTargetFile :=flag.String("targetFile","/mnt/hgfs/vmware_share/gwan4nodesLog/resultTransactions.txt","store analysis result to targetFile")
	*/

	flag.Parse()	//If we delete this line ,we can not see the help from './txChecker -h'

	resultCh := make(chan string, resultWriterBuffer)
	go ResultWriter(*optTargetFile, resultCh)	//writer write result into the target file.

	//change miner addresses to lower mode.
	for i:=0; i < len(nodeId);i++  {
		nodeId[i] = strings.ToLower(nodeId[i])
	}

	if(strings.Compare(*optType, OptionTypeBlock) == 0) {	//Parse block informations from log files.
		from,err := strconv.Atoi(*optFromBlock)
		if(err != nil) {
			fmt.Println(err)
			return
		}
		to,err := strconv.Atoi(*optToBlock)
		if(err != nil) {
			fmt.Println(err)
			return
		}

		ParseBlockFile(*optBlockInfoFile, from, to)
		ParseLogFileForBlocks(*optTargetFile, resultCh, *optFromBlock, *optToBlock)
	} else if(strings.Compare(*optType, optionTypeTransaction) == 0) {	//Parase transaction informations from log files.
		ParaseLogFileForTxs(*optTranFile, resultCh)  //for transactions.
	}
}
