package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davidwalter0/k8s-bldr-api/dispatch"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

type ValidateService interface {
	ValidateRequest(string) (string, error)
	DeployRequest(string) (string, error)
}

type validateService struct{}

func (validateService) ValidateRequest(request *dispatch.Request) (response string, err error) {
	if request == nil {
		return "", ErrorEmpty
	}
	log.Printf("%v\n", request)
	s := fmt.Sprintf("%v\n", request)
	return s, err
}

func main() {
	ctx := context.Background()
	svc := validateService{}

	requestHandler := httptransport.NewServer(
		ctx,
		makeValidateRequestEndpoint(svc),
		decodeValidateRequest,
		encodeResponse,
	)

	deployHandler := httptransport.NewServer(
		ctx,
		makeDeployRequestEndpoint(svc),
		decodeValidateRequest,
		encodeResponse,
	)

	http.Handle("/api/v1/validate", requestHandler)
	// http.Handle("/api/v1/validate/", requestHandler)
	//	http.Handle("/", requestHandler)
	http.Handle("/api/v1/deploy", deployHandler)
	http.Handle("/api/", deployHandler)
	log.Fatal(http.ListenAndServe(":9999", nil))
}

var exitExecError int = 3
var exitNormal int = 0

/*
ways to initialize?
	response := dispatch.Response{
		ApiVersion: apiVersion,
		Kind:       "response",
		// Spec:       dispatch.Spec{},
		Spec: dispatch.Spec{
			States: []dispatch.Status{status},
		},
		// Status:     {},
		// Spec:       {},
	}
	states := []dispatch.Status{status}
	response.Spec.States = states
*/

// func NewResponse(name string, err error, exit int) *dispatch.Response {
// 	status := dispatch.Status{Error: fmt.Sprintf("%v", err), Exit: exit, Name: name}
// 	response := dispatch.Response{
// 		ApiVersion: apiVersion,
// 		Kind:       "response",
// 		Spec:       dispatch.Spec{},
// 	}
// 	return &response
// }

func makeValidateRequestEndpoint(svc validateService) endpoint.Endpoint {
	var exit int
	return func(ctx context.Context, requestInterface interface{}) (interface{}, error) {
		request := requestInterface.(dispatch.Request)
		if *debug {
			dispatch.Info.Println(request)
			dispatch.Plain.Printf("\n------------------------------------------------------------------------\n")
			dispatch.Plain.Println(ctx)
			dispatch.Plain.Printf("\n------------------------------------------------------------------------\n")
			dispatch.Plain.Println(request)
			dispatch.Plain.Printf("\n------------------------------------------------------------------------\n")
			dispatch.Plain.Println("Service\n", svc)
			dispatch.Plain.Printf("\n------------------------------------------------------------------------\n")
			dispatch.MapRecursePrint(request)
			dispatch.Plain.Printf("------------------------------------------------------------------------\n")
		}
		v, err := svc.ValidateRequest(&request)
		if err != nil {
			if *debug {
				dispatch.Info.Println(v)
				dispatch.Info.Printf("ctx %v\n", ctx)
			}
			response := dispatch.NewResponse("API Call from /api/v1/validate")
			response.Append(request.Spec[0].Name, err, exitExecError, "")
			return response, nil
		}
		if *debug {
			dispatch.Info.Println(v)
		}
		requestResponse := dispatch.Request(request)
		response := dispatch.NewResponse("API Call from /api/v1/validate")
		exit, err, response = dispatch.ApiDispatch(&requestResponse, response.Filename, *debug, *verbose)
		return response, err
	}
}

func makeDeployRequestEndpoint(svc validateService) endpoint.Endpoint {
	var exit int
	return func(ctx context.Context, requestInterface interface{}) (interface{}, error) {
		request := requestInterface.(dispatch.Request)
		v, err := svc.ValidateRequest(&request)
		if err != nil {
			if *debug {
				dispatch.Info.Println(v)
				dispatch.Info.Printf("ctx %v\n", ctx)
			}
			response := dispatch.NewResponse("API Call from /api/v1/deploy")
			response.Append(request.Spec[0].Name, err, exitExecError, "error")
			return response, nil
		}
		if *debug {
			dispatch.Info.Println(v)
		}
		response := dispatch.NewResponse("API Call from /api/v1/deploy")
		var reader dispatch.ReadWriteStringBuffer
		err = json.NewEncoder(reader).Encode(request)
		if false {
			dispatch.Info.Println(exit)
		}
		response.Append(request.Spec[0].Name, err, exitExecError, string(reader))
		return response, nil
	}
}

func decodeValidateRequest(r *http.Request) (interface{}, error) {
	var request dispatch.Request

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}

	if *debug {
		dispatch.Info.Printf("%v\n", request)
	}
	/*
		if *debug {
			dispatch.Plain.Printf("\nRequest\n------------------------------------------------------------------------\n")
			dispatch.Plain.Printf("\n------------------------------------------------------------------------\n")
			dispatch.Plain.Println(request)
			dispatch.Plain.Printf("\n------------------------------------------------------------------------\n")
			dispatch.Plain.Println(r)

			dispatch.Info.Println(r.Body)
		}
		dispatch.Info.Println(r.URL)

		var m map[string]interface{}
		Info.Printf("%v\n", *r)
		var buffer []byte
		// json.Unmarshal([]byte(*r), &buffer)
		text := dispatch.ReadWriteStringBuffer(buffer)
		Info.Println(text)
		Info.Println(dispatch.Jsonify(*r))
		Info.Println(dispatch.Jsonify(r.Header))
		Info.Println(dispatch.Jsonify(r.Body))
		Info.Println(dispatch.Jsonify(text))
		js := json.NewDecoder(text)
		Info.Println(dispatch.Jsonify(js))
		text = dispatch.ReadWriteStringBuffer(dispatch.Jsonify(*r))
		var x interface{}
		x, _ = dispatch.Mapify(text)
		m = x.(map[string]interface{})
		dispatch.MapRecursePrint(m)

		x, _ = dispatch.Mapify(*r)
		m = x.(map[string]interface{})
		dispatch.MapRecursePrint(m)
		dispatch.MapRecursePrint(r.TransferEncoding)
		dispatch.MapRecursePrint(r.Method)

		dispatch.MapRecursePrint(r.Header)
		x, _ = dispatch.Mapify(r.Header)
		m = x.(map[string]interface{})
		dispatch.MapRecursePrint(m)

		dispatch.MapRecursePrint(r.Body)
		dispatch.MapRecursePrint(r.URL)
		dispatch.Info.Println(r.URL)
		dispatch.Info.Println(r.TransferEncoding[:])
		dispatch.Info.Println(r.Method)
		dispatch.Info.Println(r.Header)
		dispatch.Info.Println(r.Trailer)
		dispatch.Info.Println(r.MultipartForm)
		dispatch.Info.Println(r.Body)
		dispatch.Info.Println(r.URL)

		switch r.Method {
		case "GET":
			Info.Println("Get method will execute now")
		case "PUT":
			Info.Println("Put method will execute now")
		case "DELETE":
			Info.Println("Delete method will execute now")
		case "POST":
			Info.Println("POST method will execute now")
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return nil, err
		}

		if *debug {
			dispatch.Info.Printf("%v\n", request)
		}
	*/
	return request, nil
}

func encodeResponse(w http.ResponseWriter, response interface{}) error {
	if *debug {
		dispatch.Info.Printf("encodeResponse %v\n%v\n", w, response)
	}
	return json.NewEncoder(w).Encode(response)
}

var ErrorEmpty = errors.New("empty string")
var ErrorInvalidRequest = errors.New("Invalid Request")
