/*
*
*	Gosora Main File
*	Copyright Azareal 2016 - 2020
*
 */
// Package main contains the main initialisation logic for Gosora
package main // import "github.com/Azareal/Gosora"

import (
	"bytes"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Azareal/Gosora/common"
	"github.com/Azareal/Gosora/common/counters"
	"github.com/Azareal/Gosora/common/phrases"
	"github.com/Azareal/Gosora/query_gen"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

var router *GenRouter

// TODO: Wrap the globals in here so we can pass pointers to them to subpackages
var globs *Globs

type Globs struct {
	stmts *Stmts
}

// Experimenting with a new error package here to try to reduce the amount of debugging we have to do
// TODO: Dynamically register these items to avoid maintaining as much code here?
func afterDBInit() (err error) {
	acc := qgen.NewAcc()
	common.Rstore, err = common.NewSQLReplyStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.Prstore, err = common.NewSQLProfileReplyStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}

	err = phrases.InitPhrases(common.Site.Language)
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Loading the static files.")
	err = common.Themes.LoadStaticFiles()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.StaticFiles.Init()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.StaticFiles.JSTmplInit()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the widgets")
	common.Widgets = common.NewDefaultWidgetStore()
	err = common.InitWidgets()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the menu item list")
	common.Menus = common.NewDefaultMenuStore()
	err = common.Menus.Load(1) // 1 = the default menu
	if err != nil {
		return errors.WithStack(err)
	}
	menuHold, err := common.Menus.Get(1)
	if err != nil {
		return errors.WithStack(err)
	}
	fmt.Printf("menuHold: %+v\n", menuHold)
	var b bytes.Buffer
	menuHold.Build(&b, &common.GuestUser, "/")
	fmt.Println("menuHold output: ", string(b.Bytes()))

	log.Print("Initialising the authentication system")
	common.Auth, err = common.NewDefaultAuth()
	if err != nil {
		return errors.WithStack(err)
	}

	log.Print("Initialising the stores")
	common.WordFilters, err = common.NewDefaultWordFilterStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.MFAstore, err = common.NewSQLMFAStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.Pages, err = common.NewDefaultPageStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.Reports, err = common.NewDefaultReportStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.Emails, err = common.NewDefaultEmailStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.LoginLogs, err = common.NewLoginLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.RegLogs, err = common.NewRegLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.ModLogs, err = common.NewModLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.AdminLogs, err = common.NewAdminLogStore(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	common.IPSearch, err = common.NewDefaultIPSearcher()
	if err != nil {
		return errors.WithStack(err)
	}
	if common.Config.Search == "" || common.Config.Search == "sql" {
		common.RepliesSearch, err = common.NewSQLSearcher(acc)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	common.Subscriptions, err = common.NewDefaultSubscriptionStore()
	if err != nil {
		return errors.WithStack(err)
	}
	common.Attachments, err = common.NewDefaultAttachmentStore()
	if err != nil {
		return errors.WithStack(err)
	}
	common.Polls, err = common.NewDefaultPollStore(common.NewMemoryPollCache(100)) // TODO: Max number of polls held in cache, make this a config item
	if err != nil {
		return errors.WithStack(err)
	}
	common.TopicList, err = common.NewDefaultTopicList()
	if err != nil {
		return errors.WithStack(err)
	}
	common.PasswordResetter, err = common.NewDefaultPasswordResetter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	// TODO: Let the admin choose other thumbnailers, maybe ones defined in plugins
	common.Thumbnailer = common.NewCaireThumbnailer()

	log.Print("Initialising the view counters")
	counters.GlobalViewCounter, err = counters.NewGlobalViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.AgentViewCounter, err = counters.NewDefaultAgentViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.OSViewCounter, err = counters.NewDefaultOSViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.LangViewCounter, err = counters.NewDefaultLangViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.RouteViewCounter, err = counters.NewDefaultRouteViewCounter(acc)
	if err != nil {
		return errors.WithStack(err)
	}
	counters.PostCounter, err = counters.NewPostCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.TopicCounter, err = counters.NewTopicCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.TopicViewCounter, err = counters.NewDefaultTopicViewCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.ForumViewCounter, err = counters.NewDefaultForumViewCounter()
	if err != nil {
		return errors.WithStack(err)
	}
	counters.ReferrerTracker, err = counters.NewDefaultReferrerTracker()
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// TODO: Split this function up
func main() {
	// TODO: Recover from panics
	/*defer func() {
		r := recover()
		if r != nil {
			log.Print(r)
			debug.PrintStack()
			return
		}
	}()*/
	common.StartTime = time.Now()

	// TODO: Have a file for each run with the time/date the server started as the file name?
	// TODO: Log panics with recover()
	f, err := os.OpenFile("./logs/ops-"+strconv.FormatInt(common.StartTime.Unix(), 10)+".log", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	common.LogWriter = io.MultiWriter(os.Stderr, f)
	log.SetOutput(common.LogWriter)
	log.Print("Running Gosora v" + common.SoftwareVersion.String())
	fmt.Println("")

	// TODO: Add a flag for enabling the profiler
	if false {
		f, err := os.Create("./logs/cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	jsToken, err := common.GenerateSafeString(80)
	if err != nil {
		log.Fatal(err)
	}
	common.JSTokenBox.Store(jsToken)

	log.Print("Loading the configuration data")
	err = common.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Processing configuration data")
	err = common.ProcessConfig()
	if err != nil {
		log.Fatal(err)
	}

	err = common.InitTemplates()
	if err != nil {
		log.Fatal(err)
	}
	common.Themes, err = common.NewThemeList()
	if err != nil {
		log.Fatal(err)
	}
	common.TopicListThaw = common.NewSingleServerThaw()

	err = InitDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	buildTemplates := flag.Bool("build-templates", false, "build the templates")
	flag.Parse()
	if *buildTemplates {
		err = common.CompileTemplates()
		if err != nil {
			log.Fatal(err)
		}
		err = common.CompileJSTemplates()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err = afterDBInit()
	if err != nil {
		log.Fatalf("%+v", err)
	}

	err = common.VerifyConfig()
	if err != nil {
		log.Fatal(err)
	}

	if !common.Dev.NoFsnotify {
		log.Print("Initialising the file watcher")
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		go func() {
			var modifiedFileEvent = func(path string) error {
				var pathBits = strings.Split(path, "\\")
				if len(pathBits) == 0 {
					return nil
				}
				if pathBits[0] == "themes" {
					var themeName string
					if len(pathBits) >= 2 {
						themeName = pathBits[1]
					}
					if len(pathBits) >= 3 && pathBits[2] == "public" {
						// TODO: Handle new themes freshly plopped into the folder?
						theme, ok := common.Themes[themeName]
						if ok {
							return theme.LoadStaticFiles()
						}
					}
				}
				return nil
			}

			// TODO: Expand this to more types of files
			var err error
			for {
				select {
				case event := <-watcher.Events:
					// TODO: Handle file deletes (and renames more graciously by removing the old version of it)
					if event.Op&fsnotify.Write == fsnotify.Write {
						log.Println("modified file:", event.Name)
						err = modifiedFileEvent(event.Name)
					} else if event.Op&fsnotify.Create == fsnotify.Create {
						log.Println("new file:", event.Name)
						err = modifiedFileEvent(event.Name)
					} else {
						log.Println("unknown event:", event)
						err = nil
					}
					if err != nil {
						common.LogError(err)
					}
				case err = <-watcher.Errors:
					common.LogWarning(err)
				}
			}
		}()

		// TODO: Keep tabs on the (non-resource) theme stuff, and the langpacks
		err = watcher.Add("./public")
		if err != nil {
			log.Fatal(err)
		}
		err = watcher.Add("./templates")
		if err != nil {
			log.Fatal(err)
		}
		for _, theme := range common.Themes {
			err = watcher.Add("./themes/" + theme.Name + "/public")
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	log.Print("Initialising the task system")

	// Thumbnailer goroutine, we only want one image being thumbnailed at a time, otherwise they might wind up consuming all the CPU time and leave no resources left to service the actual requests
	// TODO: Could we expand this to attachments and other things too?
	thumbChan := make(chan bool)
	go common.ThumbTask(thumbChan)
	go tickLoop(thumbChan)

	// Resource Management Goroutine
	go func() {
		ucache := common.Users.GetCache()
		tcache := common.Topics.GetCache()
		if ucache == nil && tcache == nil {
			return
		}

		var lastEvictedCount int
		var couldNotDealloc bool
		var secondTicker = time.NewTicker(time.Second)
		for {
			select {
			case <-secondTicker.C:
				// TODO: Add a LastRequested field to cached User structs to avoid evicting the same things which wind up getting loaded again anyway?
				if ucache != nil {
					ucap := ucache.GetCapacity()
					if ucache.Length() <= ucap || common.Users.GlobalCount() <= ucap {
						couldNotDealloc = false
						continue
					}
					lastEvictedCount = ucache.DeallocOverflow(couldNotDealloc)
					couldNotDealloc = (lastEvictedCount == 0)
				}
			}
		}
	}()

	log.Print("Initialising the router")
	router, err = NewGenRouter(http.FileServer(http.Dir("./uploads")))
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Initialising the plugins")
	common.InitPlugins()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		// TODO: Gracefully shutdown the HTTP server
		runTasks(common.ShutdownTasks)
		common.StoppedServer("Received a signal to shutdown: ", sig)
	}()

	// Start up the WebSocket ticks
	common.WsHub.Start()

	if false {
		f, err := os.Create("./logs/cpu.prof")
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
	}

	//if profiling {
	//	pprof.StopCPUProfile()
	//}
	startServer()
	args := <-common.StopServerChan
	if false {
		pprof.StopCPUProfile()
		f, err := os.Create("./logs/mem.prof")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		runtime.GC()
		err = pprof.WriteHeapProfile(f)
		if err != nil {
			log.Fatal(err)
		}
	}
	// Why did the server stop?
	log.Fatal(args...)
}

func startServer() {
	// We might not need the timeouts, if we're behind a reverse-proxy like Nginx
	var newServer = func(addr string, handler http.Handler) *http.Server {
		rtime := common.Config.ReadTimeout
		if rtime == 0 {
			rtime = 8
		} else if rtime == -1 {
			rtime = 0
		}
		wtime := common.Config.WriteTimeout
		if wtime == 0 {
			wtime = 10
		} else if wtime == -1 {
			wtime = 0
		}
		itime := common.Config.IdleTimeout
		if itime == 0 {
			itime = 120
		} else if itime == -1 {
			itime = 0
		}
		return &http.Server{
			Addr:    addr,
			Handler: handler,

			ReadTimeout:  time.Duration(rtime) * time.Second,
			WriteTimeout: time.Duration(wtime) * time.Second,
			IdleTimeout:  time.Duration(itime) * time.Second,

			TLSConfig: &tls.Config{
				PreferServerCipherSuites: true,
				CurvePreferences: []tls.CurveID{
					tls.CurveP256,
					tls.X25519,
				},
			},
		}
	}

	// TODO: Let users run *both* HTTP and HTTPS
	log.Print("Initialising the HTTP server")
	if !common.Site.EnableSsl {
		if common.Site.Port == "" {
			common.Site.Port = "80"
		}
		log.Print("Listening on port " + common.Site.Port)
		go func() {
			common.StoppedServer(newServer(":"+common.Site.Port, router).ListenAndServe())
		}()
		return
	}

	if common.Site.Port == "" {
		common.Site.Port = "443"
	}
	if common.Site.Port == "80" || common.Site.Port == "443" {
		// We should also run the server on port 80
		// TODO: Redirect to port 443
		go func() {
			log.Print("Listening on port 80")
			common.StoppedServer(newServer(":80", &HTTPSRedirect{}).ListenAndServe())
		}()
	}
	log.Printf("Listening on port %s", common.Site.Port)
	go func() {
		common.StoppedServer(newServer(":"+common.Site.Port, router).ListenAndServeTLS(common.Config.SslFullchain, common.Config.SslPrivkey))
	}()
}
