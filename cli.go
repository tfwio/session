package session

import "fmt"

// CLIList returns a  sessions for CLI.
// The method first fetches a list of User elements
// then reports the Sessions with user-data (name).
func CLIList() []Session {
	var users []User
	var sessions []Session
	db, err := iniC("error(session-cli-list) loading db\n")

	defer db.Close()
	if !err {
		usermap := UserGetList()
		for m, x := range users {
			fmt.Printf("--> %04d: %s\n", m, x.Name)
			usermap[x.ID] = x
		}
		// list sessions
		db.Find(&sessions)
		fmt.Printf("--> found %d entries\n", len(sessions))
		for _, x := range sessions {
			fmt.Printf("--> '%s'\n  CRD: %s\n  EXP: %s\n  SID: %s\n",
				usermap[x.UserID].Name,
				x.Created.Format("20060102_1504.005"),
				x.Expires.Format("20060102_1504.005"),
				x.SessID)
		}
	}
	return sessions
}
