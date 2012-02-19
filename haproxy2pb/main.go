package main

import (
	"os"
	"bufio"
	"haproxy"
	"fmt"
  proto "code.google.com/p/goprotobuf/proto"
)

func readLine(r *bufio.Reader) ([]byte, error) { 
  line, isPrefix, err := r.ReadLine() 
  if !isPrefix { 
    return line, err 
  } 
  buf := append([]byte(nil), line...) 
  for isPrefix && err == nil { 
    line, isPrefix, err = r.ReadLine() 
    buf = append(buf, line...) 
  } 
  return buf, err 
} 

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		line, err := readLine(reader)
		if err != nil {
			break
		}

		pb := haproxy.Request{}
		if err = haproxy.Scan(string(line), &pb); err == nil {
			buf, err := proto.Marshal(&pb)
			if err != nil {
				fmt.Println("err", err)
			}
			fmt.Println(len(buf), len(line), len(line) - len(buf))
			//fmt.Println(pb.String())
			//fmt.Println("ok", *pb.Timestamp)
		} else {
			fmt.Println("err", err, string(line))
		}
	}

	println("done")
}
