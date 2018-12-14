package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/only1isus/majorProj/control"
)

func welcome() {
	version := "1.1.0"
	fmt.Printf(`										
                            |/
                        . - |~ .
                     .*          ', 
                    ,   #          '
                    '              *		
                    '              '  
					'-,___.'__,.-'
					==============================================
					AMBOROSA - alpha
					%v
					==============================================
					
					`, version)
}

func main() {
	c := make(chan os.Signal, 1)
	kill := make(chan bool)
	welcome()
	public := flag.Bool("public", false, "if set to true then the system will make the info public")
	if *public {
		fmt.Println("The system is public.")
	}

	d := control.Devices{}
	fan, err := d.Get("fan")
	if err != nil {
		fmt.Println(err)
	}
	// defer fan.GracefulKill(kill)
	t := control.Temperature{}
	if err := t.Maintain(30, fan); err != nil {
		fmt.Println(err)
	}

	// cleaning up. Closing all pins and turning off all devices whenever Ctrl ^ C is recieved.
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		kill <- true
	}()
	select {
	case _ = <-kill:
		fan.Off()
		fmt.Println("cleaned!")
		os.Exit(0)
	}

	// go server.Serve()
}
