// amt - program to access an SQLite database and lookup acronyms
//
// author:	Simon Rowe <simon@wiremoons.com>
// license: open-source released under The MIT License (MIT).

package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// SET GLOBAL VARIABLES

// set the version of the app here
var appversion = "0.5.4"
var appname string

// below are the flag variables used for command line args
var dbName string
var searchTerm string
var wildLookUp bool
var debugSwitch bool
var helpMe bool
var addNew bool
var showVer bool

// used to keep track of database record count
var recCount int64

// used to hold any errors
var err error

// create a global db handle - so can be used across functions
var db *sql.DB

// init always runs before applications main() function and is used here to
// set-up the required 'flag' variables from the command line parameters
// provided by the user when they run the app.
func init() {
	// flag types available are: IntVar; StringVar; BoolVar
	// flag parameters are: variable, cmd line flag, initial value, description
	// description is used by flag.Usage() on error or for help output
	flag.StringVar(&dbName, "f", "", "\tprovide SQLite database `filename` and path")
	flag.StringVar(&searchTerm, "s", "", "\t`acronym` to search for")
	flag.BoolVar(&wildLookUp, "w", false, "\tsearch for any similar matches")
	flag.BoolVar(&debugSwitch, "d", false, "\tshow debug output")
	flag.BoolVar(&helpMe, "h", false, "\tdisplay help for this program")
	flag.BoolVar(&showVer, "v", false, "\tdisplay program version")
	flag.BoolVar(&addNew, "n", false, "\tadd a new acronym record")
	// get the command line args passed to the program
	flag.Parse()
	// get the name of the application as called from the command line
	appname = filepath.Base(os.Args[0])
}

// main is the application start up function for amt
func main() {

	// confirm that debug mode is enabled and display other command
	// line flags and their current status for confirmation
	if debugSwitch {
		log.Println("DEBUG: Debug mode enabled")
		log.Printf("DEBUG: Number of command line arguments set by user is: %d", flag.NFlag())
		log.Printf("DEBUG: Command line argument settings are:")
		log.Println("\t\tDatabase name to use via command line:", dbName)
		log.Println("\t\tAcronym to search for:", searchTerm)
		log.Println("\t\tLook for similar matches:", strconv.FormatBool(wildLookUp))
		log.Println("\t\tDisplay additional debug output when run:", strconv.FormatBool(debugSwitch))
		log.Println("\t\tDisplay additional help information:", strconv.FormatBool(helpMe))
		log.Println("\t\tAdd a new acronym record:", strconv.FormatBool(addNew))
		log.Println("\t\tShow the applications version:", strconv.FormatBool(addNew))
	}

	// a function that will run at the end of the program
	defer func() {
		// END OF MAIN()
		fmt.Printf("\nAll is well\n")
	}()

	// override Go standard flag.Usage() function to get better
	// formating and output by using my own function instead
	flag.Usage = func() {
		if debugSwitch {
			log.Println("DEBUG: Running flag.Usage override function")
		}
		myUsage()
	}

	// print out start up banner
	if debugSwitch {
		log.Println("DEBUG: Calling 'printBanner()'")
	}
	printBanner()

	// check if a valid database file has been provided - either via the
	// environment variable $ACRODB or via the command line from the user
	checkDB()

	// check if a valid database file is available on the system
	if debugSwitch {
		log.Println("DEBUG: Calling 'checkDB()'")
	}
	err = checkDB()
	if err != nil {
		log.Fatal(err)
	}

	// open the database and retrive initial and print to screen
	if debugSwitch {
		log.Println("DEBUG: Calling 'openDB()'")
	}

	// open the database - or abort if fails get handle to database
	// file as 'db' for future use
	db, err = sql.Open("sqlite3", dbName)
	err = openDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// check the connection to database is ok
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Database connection status:  √")

	// display the SQLite database version we are compiled with
	fmt.Printf("SQLite3 Database Version:  %s\n", sqlVersion())
	// get current record count into global var for future use
	recCount = checkCount()
	// display the current number acronym records held in the database
	fmt.Printf("Current record count is:  %s\n", humanize.Comma(recCount))
	// display last acronym entered in the database for info
	fmt.Printf("Last acronym entered was:  '%s'\n", lastAcronym())

	if debugSwitch {
		log.Println("DEBUG: Start 'switch'...")
	}

	switch {
	case helpMe:
		flag.Usage()
		versionInfo()
		return
	case showVer:
		versionInfo()
		return
	case addNew:
		addRecord()
		return
	case len(searchTerm) > 0:
		searchRecord()
		return
	default:
		if debugSwitch {
			log.Println("DEBUG: Default switch statement called")
		}
		versionInfo()
		flag.Usage()
		return
	}

	// PROGRAM END

}
