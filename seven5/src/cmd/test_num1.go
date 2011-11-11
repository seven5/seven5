package main

import (
	"seven5"
	"fmt"
	"os"
)

func main() {

	me,err:=seven5.Type4UUID()
	if err!=nil {
		fmt.Fprintf(os.Stderr,"unable to get new UUID:%s\n",err)
	}
	
	mongrel2:=seven5.NewMongrel2("tcp://127.0.0.1:9997","tcp://127.0.0.1:9996",me)
	
	//read a single request
	req, err:=mongrel2.ReadMessage()	
	if err!=nil {
		fmt.Fprintf(os.Stderr,"error reading message:%s\n",err)
		return
	}
	fmt.Printf("server %s sent %s from client %d\n",req.ServerId, req.Path, req.ClientId)
	
	//create a response
	response:=new(seven5.Response) 
	response.UUID=req.ServerId
	response.Client=[]int{req.ClientId}
	response.Body=fmt.Sprintf("<pre>hello there, %s with client %d!</pre>",req.ServerId, req.ClientId)
	
	//send it
	err=mongrel2.WriteMessage(response)
	if err!=nil {
		fmt.Fprintf(os.Stderr,"error reading message:%s\n",err)
		return
	}
}
