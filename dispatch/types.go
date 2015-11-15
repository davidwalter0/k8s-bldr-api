package dispatch

import (
	"fmt"
	strftime "github.com/jehiah/go-strftime"
	"time"
)

func TimeStamp() (TimeStamp string) {
	TimeStamp = strftime.Format("%Y.%m.%d.%H.%M.%S.%z", time.Now())
	return
}

type Request struct {
	ApiVersion string `json:"apiVersion"` // v0.1
	Kind       string `json:"kind"`       // test
	Spec       []struct {
		Name string `json:"name"` // test-name
		Send struct {
			Spec struct {
				Arg []string          `json:"arg"` // Hello World!
				Env map[string]string `json:"env"`
				URI string            `json:"uri"` // file:///bin/echo
			} `json:"spec"`
		} `json:"send"`
		Recv struct {
			Base64 bool              `json:"base64"`
			Ack    map[string]string `json:"ack"`
			Nak    map[string]string `json:"nak"`
		} `json:"recv"`
	} `json:"spec"`
}

// type State struct {
//     id string `json:"id" bson:"id"`
//     Cities []City
// }

// type City struct {
//     id string `json:"id" bson:"id"`
// }

// func main(){
//     state := State {
//         id: "CA",
//         Cities:  []City{
//              {"SF"},
//         },
//     }

//     fmt.Println(state)
// }

type Status struct {
	TimeStamp string `json:"timestamp"`
	Error     string `json:"error"`  // extended text
	Exit      int    `json:"exit"`   // 0
	Name      string `json:"name"`   // k8s-healthz-api-reply
	Text      string `json:"text"`   // k8s-healthz-api-reply
	Output    string `json:"output"` // k8s-healthz-api-reply
}

type Spec struct {
	States []Status `json:"status"`
}
type Response struct {
	Filename   string `json:"filename"`
	ApiVersion string `json:"apiVersion"` // v0.1
	Kind       string `json:"kind"`       // result
	TimeStamp  string `json:"timestamp"`
	Spec       `json:"spec"`
}

func NewResponse(filename string) *Response {
	response := Response{
		Filename:   filename,
		TimeStamp:  TimeStamp(),
		ApiVersion: *ApiVersion,
		Kind:       "response",
		Spec:       Spec{},
	}
	return &response
}

func (response *Response) Append(name string, err error, exit int, text, output string) {
	Info.Println(name, err, exit, text, output)
	status := Status{Error: fmt.Sprintf("%v", err), Exit: exit, Name: name, TimeStamp: TimeStamp(), Text: text, Output: output}
	response.Spec.States = append(response.Spec.States, status)
}
