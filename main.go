package main

import (
	"flag"
	"fmt"
	"time"

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
	welcome()
	public := flag.Bool("public", false, "if set to true then the system will make the info public")
	if *public {
		fmt.Println("The system is public.")
	}

	fan := control.Fan{}
	fan.On()
	time.Sleep(3 * time.Second)
	fan.Off()
}
