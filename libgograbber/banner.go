package libgograbber

import (
	"fmt"
	"reflect"
	"strings"
)

// PrintBanner prints the banner... HOW GOOD IS THE BANNER?
func PrintBanner(s *State) {
	var banner string
	banner = `                                                 $$\       $$\                           
                                                 $$ |      $$ |                          
 $$$$$$\   $$$$$$\   $$$$$$\   $$$$$$\  $$$$$$\  $$$$$$$\  $$$$$$$\   $$$$$$\   $$$$$$\  
$$  __$$\ $$  __$$\ $$  __$$\ $$  __$$\ \____$$\ $$  __$$\ $$  __$$\ $$  __$$\ $$  __$$\ 
$$ /  $$ |$$ /  $$ |$$ /  $$ |$$ |  \__|$$$$$$$ |$$ |  $$ |$$ |  $$ |$$$$$$$$ |$$ |  \__|
$$ |  $$ |$$ |  $$ |$$ |  $$ |$$ |     $$  __$$ |$$ |  $$ |$$ |  $$ |$$   ____|$$ |      
\$$$$$$$ |\$$$$$$  |\$$$$$$$ |$$ |     \$$$$$$$ |$$$$$$$  |$$$$$$$  |\$$$$$$$\ $$ |      
 \____$$ | \______/  \____$$ |\__|      \_______|\_______/ \_______/  \_______|\__|      
$$\   $$ |          $$\   $$ |                                                           
\$$$$$$  |          \$$$$$$  |                                                           
 \______/            \______/`

	var version = s.Version
	var author = "Clinton \"swarley\" Carpene (@swarley777)"

	if s.VerbosityLevel > 0 {
		// g := color.New(color.FgGreen, color.Bold)

		fmt.Printf("%v\n", strings.Replace(banner, "$", g.Sprintf("$"), -1))
		fmt.Printf("%v\n", LeftPad2Len(fmt.Sprintf("Author: %v", author), " ", 89))
		fmt.Printf("%v\n", LeftPad2Len(fmt.Sprintf("Version: %v", version), " ", 89))
	}

}

func LineSep() string {
	return fmt.Sprintf("%v\n", LeftPad2Len("*", "*", 89))
}

func PrintOpts(s *State) {
	fmt.Print(LineSep())
	keys := reflect.ValueOf(s).Elem()
	typeOfT := keys.Type()
	if s.VerbosityLevel > 4 {
		for i := 0; i < keys.NumField(); i++ {
			f := keys.Field(i)
			s.Log.Debug.Printf("%s: = %v\n", typeOfT.Field(i).Name, f.Interface())
		}
		fmt.Print(LineSep())
	}

}
