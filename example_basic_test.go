package logrus_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

var _ = bytes.Compare
var _ = io.MultiWriter

func Example_wataash_logrus() {
	l := logrus.New()

	// TODO: want:
	//         json -> file
	//         text -> tty
	// ref:
	//   https://github.com/Sirupsen/logrus/issues/43
	//   https://github.com/Sirupsen/logrus/issues/230
	//   https://github.com/sirupsen/logrus/issues/673

	// affects TextFormatter.isTerminal
	// l.Out = os.Stdout // pipe
	l.Out = os.Stderr                                  // tty
	var buf bytes.Buffer                               // var buf strings.Builder
	l.Out = io.MultiWriter(os.Stdout, os.Stderr, &buf) // pipe (even stderr is)

	// pipe: time="2019-03-11T23:23:44+09:00" level=info msg=foo
	// tty:  INFO[0001] foo
	//                    ^^^^ seconds
	l.Info("foo")

	// time="2019-04-16T18:08:37+09:00" level=warning msg=foo animal=walrus number=0
	// WARN[0003] foo    animal=walrus number=0
	l.WithFields(logrus.Fields{
		"animal": "walrus",
		"number": 0,
	}).Warn("foo")

	// ignore TextFormatter.isTerminal
	l.Formatter = &logrus.TextFormatter{ForceColors: true}
	l.Info("foo") // tty and pipe: INFO[0000] foo

	// ReportCaller
	// INFO/home/.../example_basic_test.go:35 github.com/sirupsen/logrus_test.ExampleWataashLogrus() foo
	l.ReportCaller = true   // not thread safe
	l.SetReportCaller(true) // thread safe
	l.Formatter = &logrus.TextFormatter{DisableTimestamp: true}
	l.Info("foo")

	// CallerPrettyfier
	// INFO--baz-- ++bar++ foo
	l.Formatter.(*logrus.TextFormatter).CallerPrettyfier =
		func(f *runtime.Frame) (function string, file string) {
			return "++bar++", "--baz--"
		}
	l.Info("foo")

	// INFOexample_basic_test.go:99 Example_wataash_logrus foo
	//     ^ TODO: want space...
	//             inserting spece in CallerPrettyfier:file is bad idea;
	//             it's structurally wrong. (file=" example_basic_test.go:59" is wrong)
	l.Formatter = &logrus.TextFormatter{
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			// ss := strings.Split(f.Function, "/")
			// function = ss[len(ss)-1] // logrus_test.ExampleCustomFormatter
			ss := strings.Split(f.Function, ".")
			function = ss[len(ss)-1] // ExampleCustomFormatter
			file = path.Base(f.File)
			// file = fmt.Sprintf("%s:%d", file, f.Line) // example_custom_caller_test.go:49
			file = fmt.Sprintf("%s:%d", file, 99) // example_custom_caller_test.go:99
			return function, file
		},
		DisableTimestamp: true,
	}
	l.Info("foo")

	// // Output:
}

func Example_basic() {
	// os.Stdout = os.Stderr

	var log = logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	log.Formatter = new(logrus.TextFormatter)                     // default
	log.Formatter.(*logrus.TextFormatter).DisableColors = true    // remove colors
	log.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
	log.Level = logrus.TraceLevel
	log.Out = os.Stdout

	// log.Info("foo")

	// file, err := os.OpenFile("logrus.log", os.O_CREATE|os.O_WRONLY, 0666)
	// if err == nil {
	// 	log.Out = file
	// } else {
	// 	log.Info("Failed to log to file, using default stderr")
	// }

	defer func() {
		err := recover()
		if err != nil {
			entry := err.(*logrus.Entry)
			log.WithFields(logrus.Fields{
				"omg":         true,
				"err_animal":  entry.Data["animal"],
				"err_size":    entry.Data["size"],
				"err_level":   entry.Level,
				"err_message": entry.Message,
				"number":      100,
			}).Error("The ice breaks!") // or use Fatal() to force the process to exit with a nonzero code
		}
	}()

	// panic
	// log.Formatter.(*logrus.JSONFormatter).DisableTimestamp = true

	// log.Formatter = new(logrus.TextFormatter)
	// log.Trace("Went to the beach")
	// tmp := log.WithFields(logrus.Fields{
	// 	"animal": "walrus",
	// 	"number": 0,
	// })
	// tmp.Trace("Went to the beach")

	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"number": 0,
	}).Trace("Went to the beach")

	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"number": 8,
	}).Debug("Started observing beach")

	log.WithFields(logrus.Fields{
		"animal": "walrus",
		"size":   10,
	}).Info("A group of walrus emerges from the ocean")

	log.WithFields(logrus.Fields{
		"omg":    true,
		"number": 122,
	}).Warn("The group's number increased tremendously!")

	log.WithFields(logrus.Fields{
		"temperature": -4,
	}).Debug("Temperature changes")

	log.WithFields(logrus.Fields{
		"animal": "orca",
		"size":   9009,
	}).Panic("It's over 9000!")

	// Output:
	// level=trace msg="Went to the beach" animal=walrus number=0
	// level=debug msg="Started observing beach" animal=walrus number=8
	// level=info msg="A group of walrus emerges from the ocean" animal=walrus size=10
	// level=warning msg="The group's number increased tremendously!" number=122 omg=true
	// level=debug msg="Temperature changes" temperature=-4
	// level=panic msg="It's over 9000!" animal=orca size=9009
	// level=error msg="The ice breaks!" err_animal=orca err_level=panic err_message="It's over 9000!" err_size=9009 number=100 omg=true
}
