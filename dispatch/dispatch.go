package dispatch

import (
	"strings"
)

func Dispatch(filename string, debug bool, verbose bool) (exit int, err error, response *Response) {
	defer RecoverWithMessage("ManageCall", *ExitOnException, *FailureExitCode)
	unit, err := File2ReqRspInterface(filename, verbose)
	return ApiDispatch(unit, filename, debug, verbose)
}

/*
Process a json formatted request command response
*/

func ApiDispatch(unit *Request, filename string, debug bool, verbose bool) (exit int, err error, response *Response) {
	defer RecoverWithMessage("ProcessData", *ExitOnException, *FailureExitCode)
	var rc int = 1
	var version string
	response = NewResponse(filename)
	if unit != nil {
		for i := 0; i < len(unit.Spec); i++ {
			rc = 1
			var spliton string = "://"
			var split, args []string
			var uri, name, arg0, typ, expect string
			name = unit.Spec[i].Name
			var message string = "nak"
			var base64Encoded, implemented bool
			var text string

			if len(unit.Spec[i].Send.Spec.URI) > 0 {
				uri = unit.Spec[i].Send.Spec.URI
				args = unit.Spec[i].Send.Spec.Arg
				split = strings.Split(uri, spliton)
				if len(split) == 2 {
					typ, arg0 = split[0], split[1]
					if verbose {
						Info.Printf("ApiVersion [%8s] test [%-32s] execute type [%s] cmd[%s] with args [%q]\n", version, name, typ, arg0, args)
					}
					if unit.Spec[i].Recv.Base64 {
						base64Encoded = true
					}
					switch typ {
					case "http":
						implemented = true
						expect = unit.Spec[i].Recv.Ack["recv"]
						if verbose {
							Info.Printf("ApiVersion [%8s] test [%-32s] arg0  : [%q]\n", version, name, uri)
							Info.Printf("ApiVersion [%8s] test [%-32s] execute type [%s] uri [%s] with args [%q]\n", version, name, typ, uri, args)
						}
						rc, message = ExecuteHttpCheck(name, uri, args, expect, version, base64Encoded, verbose)
					case "consul":
						Info.Printf("not implemented[%s]\n", typ)
					case "https":
						Info.Printf("not implemented[%s]\n", typ)
					case "udp":
						Info.Printf("not implemented[%s]\n", typ)
					case "tcp":
						implemented = true
						expect = unit.Spec[i].Recv.Ack["recv"]
						args := strings.Join(args, " ")
						rc, message = TcpReqRsp(arg0, args, expect)
						if debug {
							Info.Printf("ApiVersion [%8s] test [%-32s] type [%s] expected [%s] msg [%s]\n", version, name, typ, expect, message)
						}
					case "cmd":
						Info.Printf("not implemented[%s]\n", typ)
					case "file":
						implemented = true
						expect = unit.Spec[i].Recv.Ack["stdout"]
						rc, message, text = ExecuteOSCmd(name, arg0, args, expect, version, base64Encoded, verbose)
						response.Append(name, err, rc, expect, text)
						if verbose {
							Info.Printf("ApiVersion [%8s] test [%-32s] arg0  : [%q]\n", version, name, arg0)
							Info.Printf("ApiVersion [%8s] test [%-32s] execute type [%s] cmd [%s] with args [%q]\n", version, name, typ, arg0, args)
							Info.Printf("ApiVersion [%8s] test [%-32s] rc      [%d] message [%s]\n text/output:%s\n", version, name, rc, message, text)
						}
					default:
						Info.Printf("ApiVersion [%8s] test [%-32s] not implemented[%s]\n", version, name, typ)
					}
					if !implemented {
						if verbose {
							Info.Printf("ApiVersion [%8s] test [%-32s] not implemented[%s]\n", version, name, typ)
						}
						continue
					}

				} else {
					message = "failed"
					rc = 1
				}
			} else {
				message = "failed"
				rc = 1
			}
			Info.Printf("ApiVersion [%8s] test [%-32s] type [%4s] status [%6s] rc [%d] uri [%s] args %q expect [%s]\n", version, name, typ, message, rc, uri, args, expect)
			if rc > exit {
				exit = rc
			}
			response.Append(name, err, rc, expect, text)
		}
	} else {
		Elog.Printf("ApiVersion [%8s] test [%-32s]\n", "unknown", "unit file invalid")
		exit = 3
	}

	return exit, err, response
}
