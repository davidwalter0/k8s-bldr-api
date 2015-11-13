package dispatch

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/davidwalter0/logger"
	"github.com/davidwalter0/transform"
	consulapi "github.com/hashicorp/consul/api"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"strings"
)

var ExitOnException *bool
var FailureExitCode *int
var Debug *bool
var Filename *string
var Verbose *bool
var ApiVersion *string

type ApiFlags struct {
	ExitOnException *bool
	FailureExitCode *int
	Debug           *bool
	Filename        *string
	Verbose         *bool
	ApiVersion      *string
}

var version string
var kind string
var Info *log.Logger = logger.Info
var Elog *log.Logger = logger.Error
var Plain *log.Logger = logger.Plain

func (flags *ApiFlags) ImportFlags(export func()) {
	export()
	ExitOnException = flags.ExitOnException
	FailureExitCode = flags.FailureExitCode
	Debug = flags.Debug
	Filename = flags.Filename
	Verbose = flags.Verbose
	ApiVersion = flags.ApiVersion
}

func (flags *ApiFlags) ExportFlags() {
	flags.ExitOnException = ExitOnException
	flags.FailureExitCode = FailureExitCode
	flags.Debug = Debug
	flags.Filename = Filename
	flags.Verbose = Verbose
	flags.ApiVersion = ApiVersion
}

func validate(expect, got string, err error) (rc int, message string) {
	if err == nil {
		if expect != got {
			rc, message = 1, "fail"
		} else {
			rc, message = 0, "ok"
		}
	} else {
		rc, message = 1, "failed"
	}
	return rc, message
}

func ExecuteOSCmd(name, arg0 string, args []string, expect, version string, b64 bool, verbose bool) (rc int, message string) {
	defer RecoverWithMessage("ExecuteOSCmd", *ExitOnException, *FailureExitCode)
	rc, message = 1, "fail"
	cmd := exec.Command(arg0, args...)
	cmd.Stdin = strings.NewReader(name)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	result := out.String()
	if *Verbose {
		Info.Printf("ApiVersion [%8s] test [%-32s] cmd := exec.Command( %s %q", version, name, arg0, args)
		Info.Printf("ApiVersion [%8s] test [%-32s] expected [%s]\n", version, name, expect)
		if expect == result {
			Info.Printf("ApiVersion [%8s] test [%-32s] found %s expected %q\n", version, name, expect, args)
		} else {
			Info.Printf("ApiVersion [%8s] test [%-32s] found %q but, expected %q\n", version, name, args, expect)
		}
		Info.Printf("ApiVersion [%8s] test [%-32s] expect [%s] out [%s]\n", version, name, expect, out.String())
	}

	if err != nil {
		rc, message = 1, fmt.Sprintf("%s", err)
		return rc, message
	}
	return validate(expect, result, err)
}

func HttpQuery(name, send, version string, verbose bool) (result string, err error) {
	defer RecoverWithMessage("HttpQuery", *ExitOnException, *FailureExitCode)
	var contents []byte
	var recv *http.Response
	recv, err = http.Get(send)
	if err != nil {
		return result, err
	} else {
		defer recv.Body.Close()
		contents, err = ioutil.ReadAll(recv.Body)
		if err != nil {
			return result, err
		}
		result = string(contents)
		if verbose {
			Info.Printf("ApiVersion [%8s] test [%-32s] %s\n", version, name, result)
		}
	}
	return result, err
}

func ExecuteHttpCheck(name, uri string, args []string, expect, version string, base64Encoded bool, verbose bool) (rc int, message string) {
	defer RecoverWithMessage("ExecuteHttpCheck", *ExitOnException, *FailureExitCode)
	rc, message = 1, "fail"
	got, err := HttpQuery(name, uri, version, verbose)
	if base64Encoded {
		got = base64.StdEncoding.EncodeToString([]byte(got))
	}
	return validate(expect, got, err)
}

func ConsulQuery(name, uri, arg0 string, args []string, expect, version string, verbose bool) (result string, err error) {
	defer RecoverWithMessage("ConsulQuery", *ExitOnException, *FailureExitCode)
	config := consulapi.DefaultConfig()
	config.Address = arg0
	consul, err := consulapi.NewClient(config)

	var api_path string = "api/v1/customer/domain"
	var root string = "api/v1"
	kv := consul.KV()
	d := &consulapi.KVPair{Key: api_path, Value: []byte(args[0])}

	kv.Put(d, nil)

	if *Debug {
		logger.Info.Println(consul, err)
	}
	kvp, qm, error := kv.Get(api_path, nil)
	if *Debug {
		logger.Info.Println(kvp, qm, error)
	}
	if err != nil {
		logger.Info.Println(err)
		return "*error*", err
	} else {
		if *Debug {
			logger.Info.Println(string(kvp.Value))
		}
	}

	keys, qm, err := kv.Keys(root, "/", nil)
	if err == nil {
		if *Debug {

			for key := range keys {
				logger.Info.Println(qm, key)
			}
		}
	}
	return string(kvp.Value), err
}

func ExecuteConsulCheck(name, uri, arg0 string, args []string, expect, version string, base64Encoded bool, verbose bool) (rc int, message string) {
	defer RecoverWithMessage("ExecuteConsulCheck", *ExitOnException, *FailureExitCode)
	rc, message = 1, "fail"
	if *Debug {
		Info.Println(name, uri, arg0, args)
	}
	got, err := ConsulQuery(name, uri, arg0, args, expect, version, verbose)
	if base64Encoded {
		got = base64.StdEncoding.EncodeToString([]byte(got))
	}
	if *Debug {
		Info.Println(name, uri, arg0, args, expect, version, verbose)
	}
	return validate(expect, got, err)
}

func File2ReqRspInterface(file string, verbose bool) (*Request, error) {
	defer RecoverWithMessage("File2ReqRspInterface", *ExitOnException, *FailureExitCode)

	isYaml := false
	parts := strings.Split(file, ".")
	if len(parts) > 1 {
		ext := parts[len(parts)-1]
		switch ext {
		case "yaml":
			isYaml = true
		case "yml":
			isYaml = true
		}
	}

	raw, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return YamlOrJson2Interface(raw, isYaml, verbose)
}

func YamlOrJson2Interface(raw []byte, isYaml, verbose bool) (*Request, error) {
	defer RecoverWithMessage("YamlOrJson2Interface", *ExitOnException, *FailureExitCode)
	var err error
	var unit *Request
	if isYaml {
		raw, err = transform.Yaml2Json(raw)
		if err != nil {
			return nil, err
		}
	}

	unit, err = Json2Request(raw)
	if err != nil {
		return nil, err
	}

	if verbose {
		version = unit.ApiVersion
		kind = unit.Kind
		Info.Printf("Raw json %v\n\n", string(raw))
		Info.Printf("Request %v\n\n", unit)
		Info.Printf("ApiVersion [%8s] Kind [%s]\n", version, kind)
	}
	return unit, nil
}

func Json2Request(raw []byte) (*Request, error) {
	defer RecoverWithMessage("Json2Request", *ExitOnException, *FailureExitCode)
	var unit Request
	err := json.Unmarshal(json.RawMessage(raw), &unit)
	if err != nil {
		switch err.(type) {
		// ignoring syntax error for now. the test will fail, but if
		// there is a recovery it might contine to the next step after
		// logging the error
		case *json.SyntaxError:
			Info.Printf("%v\n", err)
		default:
			return nil, err
			// t := reflect.TypeOf(err)
			// fmt.Println(t)
			// Elog.Printf("%v\n", err)
			// os.Exit(3)
		}
	}
	return &unit, nil
}

func TcpReqRsp(uri, text, expect string) (rc int, message string) {
	rc, message = 1, "fail"
	defer RecoverWithMessage("TcpReqRsp", *ExitOnException, *FailureExitCode)
	var got string
	var err error
	var size int = len(expect) + 1
	var connection net.Conn

	var n int = 0
	buffer := make([]byte, size)
	connection, err = net.Dial("tcp", uri)

	defer connection.Close()
	if err != nil {
		return validate(expect, "", err)
	} else {
		n, err = connection.Write([]byte(text))
		if err == nil {
			if tcpConnection, ok := connection.(*net.TCPConn); ok {
				tcpConnection.CloseWrite()
			}
		}
		n, err = connection.Read(buffer)
		if err == nil {
			if tcpConnection, ok := connection.(*net.TCPConn); ok {
				tcpConnection.CloseRead()
			}
		}
		if err == nil {
			got = string(buffer[0:n])
		}
	}
	return validate(expect, got, err)
}
