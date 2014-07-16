package go_agent

/*
#cgo linux LDFLAGS: -Wl,-rpath,lib -Llib -lnewrelic-common -lnewrelic-collector-client -lnewrelic-transaction
#cgo darwin LDFLAGS: -r/Users/ralph/_code/go/src/source.datanerd.us/ralph/go_agent/lib -L/Users/ralph/_code/go/src/source.datanerd.us/ralph/go_agent/lib -lnewrelic-common -lnewrelic-collector-client -lnewrelic-transaction
#include <stdlib.h>
#include "include/newrelic_common.h"
#include "include/newrelic_collector_client.h"
#include "include/newrelic_transaction.h"

void newrelic_register() {
    newrelic_register_message_handler(newrelic_message_handler);
}
*/
import "C"
import "unsafe"
import "net/http"

func StartAgent(licence_key string, app_name string) {
	C.newrelic_register()
	C.newrelic_init(C.CString(licence_key),
		C.CString(app_name), C.CString("Go"), C.CString("1.1.2"))
}

func StartTransaction(name string) C.long {
	var c_name *C.char = C.CString(name)
	defer C.free(unsafe.Pointer(c_name))

	var txn_id C.long = C.newrelic_transaction_begin()
	C.newrelic_transaction_set_name(txn_id, c_name)
	C.newrelic_transaction_set_request_url(txn_id, c_name)

	return txn_id
}

func StopTransaction(txn_id C.long) {
	C.newrelic_transaction_end(txn_id)
}

type InstrumentedHandler struct {
	Name string
	Handler http.Handler
}
func (h *InstrumentedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var txn_id = StartTransaction(h.Name)
	defer StopTransaction(txn_id)
	h.Handler.ServeHTTP(w, r)
}

func InstrumentHttpHandler(name string, handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var txn_id = StartTransaction(name)
		defer StopTransaction(txn_id)
		handler(w, r)
	}
}
