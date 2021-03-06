package panel

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Azareal/Gosora/common"
)

func Backups(w http.ResponseWriter, r *http.Request, user common.User, backupURL string) common.RouteError {
	basePage, ferr := buildBasePage(w, r, &user, "backups", "backups")
	if ferr != nil {
		return ferr
	}

	if backupURL != "" {
		// We don't want them trying to break out of this directory, it shouldn't hurt since it's a super admin, but it's always good to practice good security hygiene, especially if this is one of many instances on a managed server not controlled by the superadmin/s
		backupURL = common.Stripslashes(backupURL)

		var ext = filepath.Ext("./backups/" + backupURL)
		if ext != ".sql" && ext != ".zip" {
			return common.NotFound(w, r, basePage.Header)
		}
		info, err := os.Stat("./backups/" + backupURL)
		if err != nil {
			return common.NotFound(w, r, basePage.Header)
		}
		w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))

		if ext == ".sql" {
			// TODO: Change the served filename to gosora_backup_%timestamp%.sql, the time the 	file was generated, not when it was modified aka what the name of it should be
			w.Header().Set("Content-Disposition", "attachment; filename=gosora_backup.sql")
			w.Header().Set("Content-Type", "application/sql")
		} else {
			// TODO: Change the served filename to gosora_backup_%timestamp%.zip, the time the 	file was generated, not when it was modified aka what the name of it should be
			w.Header().Set("Content-Disposition", "attachment; filename=gosora_backup.zip")
			w.Header().Set("Content-Type", "application/zip")
		}
		// TODO: Fix the problem where non-existent files aren't greeted with custom 404s on ServeFile()'s side
		http.ServeFile(w, r, "./backups/"+backupURL)
		return nil
	}

	var backupList []common.BackupItem
	backupFiles, err := ioutil.ReadDir("./backups")
	if err != nil {
		return common.InternalError(err, w, r)
	}
	for _, backupFile := range backupFiles {
		var ext = filepath.Ext(backupFile.Name())
		if ext != ".sql" {
			continue
		}
		backupList = append(backupList, common.BackupItem{backupFile.Name(), backupFile.ModTime()})
	}

	pi := common.PanelBackupPage{basePage, backupList}
	return renderTemplate("panel_backups", w, r, basePage.Header, &pi)
}
