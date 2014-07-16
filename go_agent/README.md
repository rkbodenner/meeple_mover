go_agent
========
```
import agent "github.org/gnarg/go_agent"
import "net/http"

agent.StartAgent("LICENCE_KEY", "APP_NAME")
var transaction_id = agent.StartTransaction("TXN_NAME")
// do some stuff
agent.StopTransaction(transaction_id)

http.HandleFunc("/view/", agent.InstrumentHttpHandler("TXN_NAME", viewHandler))
```
