// amt - program to access an SQLite database and lookup acronyms
//
// author:	Simon Rowe <simon@wiremoons.com>
// license: open-source released under The MIT License (MIT).
//
// Package used to manipulate the SQLite database for application 'amt'
//
// Example record of 'ACRONYMS' table in SQLite database for
// reference:
//
//   rowid 			: hidden internal sqlite record id
//   Acronym 		: 21CN
//   Definition 	: 21st Century Network
//   Description    : A new BT Plc network infrastructure consolidating
//                    multiple legacy networks into one common internet protocol
//                    platform.
//   Source 		: General ICT

package lib

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dustin/go-humanize"
	_ "github.com/mattn/go-sqlite3"
)

// create a global db handle - so can be used across functions
var DB *sql.DB

func OpenDataBase() (err error) {
	// open the database and retrieve initial data - then print to
	// screen for users benefit

	if DebugSwitch {
		log.Println("DEBUG: Calling 'openDB()'")
	}
	// open the database - or abort if fails get handle to database
	// file as 'db' for future use
	err = OpenDB()
	if err != nil {
		log.Fatal(err)
	}

	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			log.Println("ERROR: unable to close the database.")
		}
	}(DB)

	// check the connection to database is ok
	err = DB.Ping()
	if err != nil {
		panic(err.Error())
	}
	// return any errors encountered
	return err
}

// checkDB is used to verify if a valid database file name and path
// has been provided by the user.
//
// The database file name can be provided to the program via the
// command line or via an environment variable named: ACRODB.
//
// The function checks to ensure the database file name provided
// exists, obtains its size on disk and checks it file permissions.
// These items are output to stdout by the function.
//
// The checkDB function returns an error if it fails to find a valid
// database file or one that can not be opened successfully. If the
// function fails for any reason the function returns with information
// summarising the error encountered.
//
// If successful the checkDB function sets the global variable
// 'Dbname' to the valid path and file name of the SQLite database to
// be used.
func CheckDB() (err error) {
	// check if user has specified the location of the database using
	// the command line '-f' flag. Using the command line flag can
	// therefore override the environment variable setting or database
	// file located in the same directory as the executable
	if DbName == "" {
		// nothing provided via command line...
		if DebugSwitch {
			log.Print("DEBUG: No database name provided via cli - checking environment instead...")
		}

		// get contents of environment variable $ACRODB
		DbName = os.Getenv("ACRODB")

		// check if a database name and path was provided via the
		// environment variable?
		if DbName == "" {

			if DebugSwitch {
				log.Println("DEBUG: No database name provided via environment variable ACRODB")
				log.Println("DEBUG: Environment variable $ACRODB is:", os.Getenv("ACRODB"))
			}

			// final attempt to find a usable database - so look in
			// the same directory as the program executable for a file
			// called: amt-db.db

			DbName = filepath.Join(filepath.Dir(os.Args[0]), "acronyms.db")

			if DebugSwitch {
				log.Printf("DEBUG: Looking for 'acronyms.db' in same location as executable: '%s'\n", filepath.Dir(os.Args[0]))
			}

			// quick check to see if file exists - full check will be done below if it gets that far...
			_, err := os.Stat(DbName)
			if err != nil {
				// no database found - return with an error
				err = errors.New("WARNING: no database containing your acronyms can be found...\n")
				return err
			}

			if DebugSwitch {
				log.Printf("DEBUG: Database file: '%s' exists - attemping to use...\n", DbName)
			}
		}
	}

	// DbName is not empty if we got here
	if DebugSwitch {
		log.Printf("DEBUG: database filename provided is: %s", DbName)
		log.Printf("DEBUG: Checking file stats for: '%s'\n", DbName)
	}

	// check 'DbName' is valid file with os.Stats()
	fi, err := os.Stat(DbName)
	if err == nil {
		mode := fi.Mode()

		if DebugSwitch {
			log.Printf("DEBUG: checking if '%s' is a regular file with os.Stat() call\n", DbName)
		}

		// check is a regular file
		if mode.IsRegular() {
			// print out some details of the database file:
			fmt.Printf("Database location: %s\nDatabase permissions: %s     Database size: %s bytes\n\n",
				filepath.Join(filepath.Dir(DbName), fi.Name()), fi.Mode(), humanize.Comma(fi.Size()))

			if DebugSwitch {
				log.Println("DEBUG: regular file check completed ok - return to main()")
			}
			// success - we are done!
			return nil
		}
	}

	if DebugSwitch {
		log.Print("DEBUG: Exiting program as specified database file ")
		log.Printf("%s is not valid file that can be accessed", DbName)
	}
	// error found with the provided database file
	err = fmt.Errorf("ERROR: database: '%s' is not a regular file\nError returned: %v\nrun 'amt --help' for more assistance\nABORT\n", DbName, err)
	return err
}

// openDB is the function used to open the database and obtain initial
// information confirming the connection is working, the acronym
// record count in the database, and the last new acronym record
// entered.
//
// The openDB function returns an error and error message to explain
// the problem encountered, or 'nil' if no errors occurred. The
// function returns no other information as the handle to the database
// is a global variable.
func OpenDB() (err error) {

	if DebugSwitch {
		log.Printf("DEBUG: Attempting to open the database: '%s' ... ", DbName)
	}

	// open the database - or abort if fails. If successful get the
	// handle to open database file as 'db' for any future SQL calls
	DB, err = sql.Open("sqlite3", DbName)
	if err != nil {

		if DebugSwitch {
			log.Printf("DEBUG: FAILED to open %s with error: %v - will exit application\n", DbName, err)
			log.Println("DEBUG: Exit program with call to 'log.Fatal()'")
		}

		err = fmt.Errorf("FATAL ERROR: unable to get handle to SQLite database file: %s\nError is: %v\n", DbName, err)
		return err
	}

	// check connection to database is ok
	err = DB.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Database connection status:  √")

	// display the SQLite database version we are compiled with
	fmt.Printf("SQLite3 Database Version:  %s\n", SqlVersion())
	// obtain and display the current record count into global var for future use
	RecCount = CheckCount()
	fmt.Printf("Current record count is:  %s\n", humanize.Comma(RecCount))
	// display last acronym entered into the database for info
	fmt.Printf("Last acronym entered was:  '%s'\n", LastAcronym())
	// all ok - return no errors
	return nil
}

// popNewDB function used to open the database and obtain initial
// information confirming the connection is working, the acronym
// record count in the database, and the last new acronym record
// entered.
//
// The openDB function returns an error and error message to explain
// the problem encountered, or 'nil' if no errors occurred. The
// function returns no other information as the handle to the database
// is a global variable.
func PopNewDB() (err error) {

	if DebugSwitch {
		log.Printf("DEBUG: Attempting to open the database: '%s' ... ", DbName)
	}

	// open the database - or abort if fails. If successful get the
	// handle to open database file as 'db' for any future SQL calls
	DB, err = sql.Open("sqlite3", DbName)
	if err != nil {

		if DebugSwitch {
			log.Printf("DEBUG: FAILED to open %s with error: %v - will exit application\n", DbName, err)
			log.Println("DEBUG: Exit program with call to 'log.Fatal()'")
		}

		err = fmt.Errorf("FATAL ERROR: unable to get handle to SQLite database file: %s\nError is: %v\n", DbName, err)
		return err
	}

	// check connection to database is ok
	err = DB.Ping()
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("Database connection status:  √")

	// display the SQLite database version we are compiled with
	fmt.Printf("SQLite3 Database Version:  %s\n", SqlVersion())
	// obtain and display the current record count into global var for future use
	RecCount = CheckCount()
	fmt.Printf("Current record count is:  %s\n", humanize.Comma(RecCount))
	// display last acronym entered into the database for info
	fmt.Printf("Last acronym entered was:  '%s'\n", LastAcronym())
	// all ok - return no errors
	return nil
}

// CheckCount provides the current total record count in the acronym
// table. The function takes no inputs. CheckCount function returns
// the record count as an int64 variable. If an error occurs obtaining
// the record count from the database it will be printed to stderr.
func CheckCount() int64 {

	if DebugSwitch {
		log.Println("DEBUG: running record count function 'CheckCount()' ... ")
	}
	// create variable to hold returned database count of records
	var RecCount int64
	// query the database to get number of records - result out in
	// variable RecCount
	err := DB.QueryRow("select count(*) from ACRONYMS;").Scan(&RecCount)
	if err != nil {
		log.Printf("ERROR in function 'CheckCount()' with SQL QueryRow: %v\n", err)
	}
	if DebugSwitch {
		log.Printf("DEBUG: records count in table returned: %d\n", RecCount)
	}
	// return the result
	return RecCount
}

// LastAcronym obtains the last acronym entered into the acronym
// table. The LastAcronym function takes not inputs. The LastAcronym
// function returns the last acronym entered into the table as a
// string variable. If an error occurs obtaining the last acronym
// entered from the database it will be printed to stderr.
//
// SQL statement run is:
//
//	SELECT Acronym FROM acronyms Order by rowid DESC LIMIT 1;
func LastAcronym() string {

	if DebugSwitch {
		log.Println("DEBUG: Getting last entered acronym... ")
	}
	// create variable to hold returned database query
	var lastEntry string
	// query the database to get last entered acronym - result
	// returned to variable 'lastEntry'
	err := DB.QueryRow("SELECT Acronym FROM acronyms Order by rowid DESC LIMIT 1;").Scan(&lastEntry)
	if err != nil {
		log.Printf("ERROR: in function 'LastAcronym()' with SQL  QueryRow (lastEntry): %v\n", err)
	}

	if DebugSwitch {
		log.Printf("DEBUG: last acronym entry in table returned: %s\n", lastEntry)
	}
	// return the result
	return lastEntry
}

// SqlVersion provides the version of SQLite library that is being
// used by the program. The function take no parameters. The
// SqlVersion function returns a string with a version number obtained
// by running the SQLite3 statement:
//
//	SELECT SQLITE_VERSION();
func SqlVersion() string {

	if DebugSwitch {
		log.Println("DEBUG: Getting SQLite3 database version of software... ")
	}
	// create variable to hold returned database query
	var dbVer string
	// query the database to get version - result returned to variable
	// 'dbVer'
	err := DB.QueryRow("select SQLITE_VERSION();").Scan(&dbVer)
	if err != nil {
		log.Printf("ERROR: in function 'SqlVersion()' with SQL QueryRow (dbVer): %v\n", err)
	}
	if DebugSwitch {
		log.Printf("DEBUG: function 'SqlVersion()' returned value from query: %s\n", dbVer)
	}
	// return the result
	return dbVer
}

// getSources provide the current 'sources' held in the acronym table
// getSources function takes no parameters. The getSources functions
// returns a string contain a list of distinct 'source' records such
// as "General ICT"
func GetSources() string {

	if DebugSwitch {
		log.Print("DEBUG: Getting source list function... ")
	}
	// create variable to hold returned database source list
	var sourceList []string
	// query the database to extract distinct 'source' records -
	// result out in variable 'sourceList'
	rows, err := DB.Query("select distinct(source) from acronyms;")
	if err != nil {
		log.Printf("ERROR: in function 'getSources()' with SQL QueryRow (sourceList): %v\n", err)
	}
	defer rows.Close()

	var source string

	for rows.Next() {
		err = rows.Scan(&source)
		sourceList = append(sourceList, source)
		if DebugSwitch {
			log.Printf("DEBUG: Source extracted: %s\n", source)
		}
	}

	fmt.Printf("\nExisting %d acronym 'source' choices:\n\n", len(sourceList))
	for idx, source := range sourceList {
		fmt.Printf("[%d]: '%s'  ", idx, source)
	}
	fmt.Printf("\n\n")
	// ask user to choose one...
	idxChoice := GetInput("Enter a source [#] for the new acronym: ")
	idxFinal, err := strconv.Atoi(idxChoice)
	// error - could not convert to Int so just return the string as is...
	if err != nil {
		return idxChoice
	}
	// check the number entered is not greater or less than it should be
	if (idxFinal > (len(sourceList) - 1)) || (idxFinal < 0) {
		// error - entered value is out of range warn user and exit
		log.Fatalf("\n\nFATAL ERROR: The source # you entered '%d' is greater than choices of '0' to '%d' offered, or less than zero\n\n", idxFinal, len(sourceList)-1)
	}
	// return the result
	return sourceList[idxFinal]
}

// addRecord function adds a new record to the acronym table held in
// the SQLite database It does not take any parameters. It does not
// return any information, and exits the program on completion. The
// application will exit of there is an error attempting to insert the
// new record into the database.
//
// The SQL insert statement used is:
//
//	insert into ACRONYMS(Acronym, Definition, Description, Source)
//	values(?,?,?,?)
func AddRecord() {

	if DebugSwitch {
		log.Printf("DEBUG: Adding new record function... \n")
	}
	// update screen for user
	fmt.Printf("\n\nADD A NEW ACRONYM RECORD\n¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯\n")
	fmt.Printf("Note: To abort the input of a new record press keys:  Ctrl + c \n\n")
	// get new acronym from user
	acronym := GetInput("Enter the new acronym: ")
	// todo: check the acronym does not already exist - check with user to continue...
	definition := GetInput("Enter the expanded version of the new acronym: ")
	description := GetInput("Enter any description for the new acronym: ")
	// show list of sources currently used and get one from the user
	source := GetSources()
	// check the user is happy with what has been collected from them...
	fmt.Printf("\nContinue to add new acronym:\n\tACRONYM: %s\n\tEXPANDED: %s\n\tDESCRIPTION: %s\n\tSOURCE: %s\n",
		acronym, definition, description, source)

	// get current database record count
	preInsertCount := CheckCount()

	// see if user wants to continue with the
	if CheckContinue() {
		// ok - add record to the database table
		_, err := DB.Exec("insert into ACRONYMS(Acronym, Definition, Description, Source) values(?,?,?,?)",
			acronym, definition, description, source)
		if err != nil {
			log.Fatalf("FATAL ERROR inserting new acronym record: %v\n", err)
		}
		// get new database record count post insert
		newInsertCount := CheckCount()
		// inform user of difference in database record counts -
		// should be 1
		fmt.Printf("SUCCESS: %d record added to the database\n",
			newInsertCount-preInsertCount)
		// inform user of database record counts
		fmt.Printf("\nDatabase record count is: %s  [was: %s]\n",
			humanize.Comma(newInsertCount), humanize.Comma(preInsertCount))
	}

	// function complete
	return
}

// searchRecord function obtains a string from the users and search
// for it in the SQLite acronyms database. It does not take any
// parameters. It does not return any information, and exits the
// program on completion. The application will exit of there is an
// error.
//
// The SQL select statement used is:
//
//	select rowid,Acronym,Definition,Description,Source from ACRONYMS where
//	Acronym like ? ORDER BY Source;
func SearchRecord(searchTerm string) {
	// start search for an acronym - update user's screen
	fmt.Printf("\n\nSEARCH FOR AN ACRONYM RECORD\n¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯\n")
	//
	// check we have a term to search for in the acronyms database
	if DebugSwitch {
		log.Printf("DEBUG: checking for a search term ... ")
	}
	if DebugSwitch {
		log.Printf("DEBUG: search term provided: %s\n", searchTerm)
	}
	// update user that the database is open and acronym we will
	// search for in how many records:
	fmt.Printf("\nSearching for:  '%s'  across %s records - please wait...\n",
		searchTerm, humanize.Comma(RecCount))

	// flush any output to the screen
	_ = os.Stdout.Sync()

	// run a SQL query to find any matching acronyms to that provided
	// by the user
	rows, err := DB.Query("select rowid,Acronym,Definition,Description,Source from ACRONYMS where Acronym like ? ORDER BY Source;", searchTerm)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Printf("\nMatching results are:\n\n")
	for rows.Next() {
		// variables to hold returned database values - use []byte
		// instead of string to get around NULL values issue error
		// which states:
		//
		// " Scan error on column index 2: unsupported driver -> Scan
		// pair: <nil> -> *string"
		var rowid, acronym, definition, description, source []byte
		err := rows.Scan(&rowid, &acronym, &definition, &description, &source)
		if err != nil {
			log.Printf("ERROR: reading database record: %v", err)
		}
		// print the current row to screen - need string(...) as
		// values are bytes
		fmt.Printf("ID: %s\nACRONYM: '%s' is: %s.\nDESCRIPTION: %s\nSOURCE: %s\n\n",
			string(rowid), string(acronym), string(definition), string(description), string(source))
	}
	// check there were no other error while reading the database rows
	err = rows.Err()
	if err != nil {
		log.Printf("ERROR: reading database row returned: %v", err)
	}
	// function complete ok
	return
}

// RemoveRecord function is used to remove (ie delete) a record from
// the Acronyms database. The record to be removed is identified by
// its 'rowid' number. The record to be removed is first displayed to
// allow the user to check it is the correct one, and on confirmation
// the record if removed from the ACRONYMS table.
//
// The 'rowid' of the record is obtained from the user via the command
// line switch '-r'. This 'rowid' is held in the global variable
// 'rmid'. The 'rowid' is provided to the function when called as a
// string value named 'rmid'.
//
// The RemoveRecord function returns either 'nil' as an err value or
// type error, or details of any actual error that occurs when it
// runs.
//
// The SQL delete statement used is:
//
//	delete from ACRONYMS where rowid = ?;
func RemoveRecord(rmid string) (err error) {
	// start remove for an acronym - update user's screen
	fmt.Printf("\n\nREMOVE AN ACRONYM RECORD\n¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯\n")
	//
	// check we have a rowid to remove from the acronyms database table:
	if DebugSwitch {
		log.Printf("DEBUG: checking for a search term ... ")
	}
	if DebugSwitch {
		log.Printf("DEBUG: record 'rowid' to remove is: %s\n", rmid)
	}
	if rmid == "" {
		log.Println("ERROR: an 'Acronym ID' for the record to be removed needs to be provided.")
		log.Println("An acronyms record 'ID' is shown as part of the output of a valid search result.")
		err = errors.New("ERROR: empty value provided for 'Acronym ID' in record removal request")
		return err
	}

	// validate the rowid is an integer
	if _, err := strconv.ParseInt(rmid, 10, 64); err != nil {
		fmt.Printf("\nERROR: acronym ID '%v' is not a valid number.\n", rmid)
		fmt.Printf("Please provide a acronym 'ID' number for the record you want to delete from the database.\n")
		err = errors.New("unable to find integer in acronym ID value: '%s'. Error returned: '%v'")
		return err
	}

	// update user that the database is open and the acronym we will
	// remove is one of a number of records:
	fmt.Printf("\nSearching for Acronym ID:  '%s'  across %s records - please wait...\n",
		rmid, humanize.Comma(RecCount))
	// flush any output to the screen
	_ = os.Stdout.Sync()

	// run a SQL query to find the matching acronym to the 'rowid'
	// provided by the user - should return a single row result or and
	// error is there is no match to the rowid
	var rowid, acronym, definition, description, source []byte
	err = DB.QueryRow("select rowid,Acronym,Definition,Description,Source from ACRONYMS where rowid = ?",
		rmid).Scan(&rowid, &acronym, &definition, &description, &source)
	// check the results obtained are good
	switch {
	// no match found
	case err == sql.ErrNoRows:
		fmt.Printf("\nNo acronym with ID: '%s' found in the database\n", rmid)
		return err
	// unknown error returned
	case err != nil:
		fmt.Printf("\nUnable to find acronym ID '%s' as query retured: %v\n", rmid, err)
		return err
		// match found so print out results
	default:
		fmt.Printf("\nRecord match found:\n\n")
		fmt.Printf("ID: %s\nACRONYM: '%s' is: %s.\nDESCRIPTION: %s\nSOURCE: %s\n\n",
			string(rowid), string(acronym), string(definition), string(description), string(source))
		fmt.Printf("\nRemove record ID '%s' for acronym: '%s'.    ", rmid, string(acronym))
	}

	// Check with the user that the record shown above is the one they
	// want to remove, before it is actually removed from the table
	if !CheckContinue() {
		fmt.Printf("Removal of Acronym ID '%s' aborted at users request\n", rmid)
		err = errors.New("Removal of Acronym ID '%s' aborted at users request\n")
		return err
	}

	fmt.Printf("Removing Acronym ID '%s' ...\n", rmid)
	// get current database record count
	preInsertCount := CheckCount()

	// ok - remove record to the database table
	_, err = DB.Exec("delete from ACRONYMS where rowid = ?", rmid)
	if err != nil {
		log.Fatalf("FATAL ERROR removing acronym record: %v\n", err)
	}
	// get new database record count post insert
	newInsertCount := CheckCount()
	// inform user of difference in database record counts -
	// should be 1
	fmt.Printf("SUCCESS: %d record removed to the database\n",
		preInsertCount-newInsertCount)
	// inform user of database record counts
	fmt.Printf("\nDatabase record count is: %s  [was: %s]\n",
		humanize.Comma(newInsertCount), humanize.Comma(preInsertCount))

	// function complete
	return err

}
