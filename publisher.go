package publisher

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/pprof"

	"github.com/go-playground/validator/v10"
	jsoniter "github.com/json-iterator/go"
	"github.com/julienschmidt/httprouter"
)

// Publisher represents a publisher.
// To create a Publisher, you should use NewPublisher.
type Publisher struct {
	// vmessServerIndex stores the mapping relationship between vmessServerID and vmess server,
	// which is used for quick search.
	vmessServerIndex map[string]*VMessServer

	// routingRuleIndex stores the mapping relationship between routingRuleID and rules.
	routingRulesIndex map[string]*[]RoutingRule

	// subscribers represents the subscribers.
	subscribers []Subscriber

	// subscriberIndex stores the mapping relationship between subscriber's key and subscriber,
	// which is used for quick search.
	subscriberIndex map[string]*Subscriber

	// router is used to route HTTP requests.
	router *httprouter.Router
}

// NewPublisher create a Publisher.
func NewPublisher() *Publisher {
	pub := &Publisher{
		vmessServerIndex: make(map[string]*VMessServer),
		subscriberIndex:  make(map[string]*Subscriber),
	}
	router := httprouter.New()
	pub.router = router

	// Registering pprof endpoints for debugging
	router.HandlerFunc(http.MethodGet, "/debug/pprof/", pprof.Index)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/cmdline", pprof.Cmdline)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/profile", pprof.Profile)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/symbol", pprof.Symbol)
	router.HandlerFunc(http.MethodPost, "/debug/pprof/symbol", pprof.Symbol)
	router.HandlerFunc(http.MethodGet, "/debug/pprof/trace", pprof.Trace)
	router.Handler(http.MethodGet, "/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Handler(http.MethodGet, "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Handler(http.MethodGet, "/debug/pprof/heap", pprof.Handler("heap"))
	router.Handler(http.MethodGet, "/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Handler(http.MethodGet, "/debug/pprof/block", pprof.Handler("block"))
	router.Handler(http.MethodGet, "/debug/pprof/mutex", pprof.Handler("mutex"))

	// Registering publish endpoint
	logger.Info("publish servers endpoint:  /publish/:key/servers")
	router.GET("/publish/:key/servers", pub.PublishServersHandle())

	// Registering publish routing rules
	logger.Info("publish routing rules endpoint:  /publish/:key/routingRules/:rulesID")
	router.GET("/publish/:key/routingRules/:rulesID", pub.PublishRoutingRulesHandle())

	return pub
}

// LoadConfigFile load and verify the configuration file.
func (p *Publisher) LoadConfigFile(path string) error {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Errorf("Read config file (%s) error: %s", path, err.Error())
		return err
	}

	api := jsoniter.Config{
		TagKey:                        "cfg",
		ObjectFieldMustBeSimpleString: true,
		DisallowUnknownFields:         true,
	}.Froze()
	iter := jsoniter.NewIterator(api)
	iter.ResetBytes(bs)

	for {
		switch field := iter.ReadObject(); field {
		case "":
			// complete
			goto LOAD_CONFIG_COMPLETE
		case "publisher":
			// parse publisher
			iter.Skip()
			continue
		case "vmessServers":
			// parse vmessServers
			iter.ReadVal(&p.vmessServerIndex)
		case "routingRules":
			// parse routingRules
			iter.ReadVal(&p.routingRulesIndex)
		case "subscribers":
			// parse subscribers
			iter.ReadVal(&p.subscribers)
		default:
			err := errors.New("Parse config error: Unknown config field '" + field + "'")
			logger.Error(err.Error())
			return err
		}

		// any error?
		if iter.Error != nil {
			logger.Errorf("Parse config error: %s", err.Error())
			return iter.Error
		}
	}

LOAD_CONFIG_COMPLETE:
	// validate config
	val := validator.New()
	for i := range p.subscribers {
		err := val.Struct(&p.subscribers[i])
		if err != nil {
			logger.Error("validate config error: " + err.Error())
			return err
		}

		for _, serverID := range p.subscribers[i].VMessServers {
			if _, ok := p.vmessServerIndex[serverID]; !ok {
				err := fmt.Errorf("Subscriber(%s) want VMess Server(%s), but not found", p.subscribers[i].Remarks, serverID)
				logger.Error(err.Error())
				return err
			}
		}

		for _, rulesID := range p.subscribers[i].RoutingRules {
			if _, ok := p.routingRulesIndex[rulesID]; !ok {
				err := fmt.Errorf("Subscriber(%s) want Routing Rule(%s), but not found", p.subscribers[i].Remarks, rulesID)
				logger.Error(err.Error())
				return err
			}
		}
	}

	// update subscriberIndex
	for i := range p.subscribers {
		key := p.subscribers[i].Key
		if _, ok := p.subscriberIndex[key]; ok {
			// duplicate key
			err := fmt.Errorf("validate config error: duplicate subscriber key (%s) found, which is not allowed", key)
			logger.Error(err.Error())
			return err
		}
		p.subscriberIndex[key] = &p.subscribers[i]
	}

	return nil
}

// Router returns the router of Publisher.
func (p *Publisher) Router() http.Handler {
	return p.router
}

// PublishServersHandle returns a function which receives HTTP requests and
// publishes the list of servers to subscribers.
func (p *Publisher) PublishServersHandle() httprouter.Handle {
	return func(respw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		key := ps.ByName("key")
		subscriber, ok := p.subscriberIndex[key]
		if !ok {
			// invalid subscriber
			logger.Warnf("Invalid key recieved: %s, response with HTTP 404", key)
			http.NotFound(respw, req)
			return
		}

		buf := &bytes.Buffer{}
		b64Writer := base64.NewEncoder(base64.StdEncoding, buf)

		for _, vmessServerID := range subscriber.VMessServers {
			server, ok := p.vmessServerIndex[vmessServerID]
			if !ok {
				// config error, should not happen
				logger.Errorf("Subscriber(%s) want VMess Server(%s), but not found, response with HTTP 500", subscriber.Remarks, vmessServerID)
				respw.WriteHeader(http.StatusInternalServerError)
				respw.Write([]byte("Server Internal Error, Please tell the administrator."))
				return
			}
			if err := server.WriteShareLink(b64Writer); err != nil {
				// marshal error
				logger.Errorf("Marshal VMess server (%s) error: %s", vmessServerID, err.Error())
				respw.WriteHeader(http.StatusInternalServerError)
				respw.Write([]byte("Server Internal Error, Please tell the administrator."))
				return
			}
			b64Writer.Write([]byte{0x0A})
		}

		b64Writer.Close()
		respw.WriteHeader(http.StatusOK)
		respw.Write(buf.Bytes())
		logger.Debugf("Success reponse to Subscriber(%s)", subscriber.Remarks)
	}
}

// PublishRoutingRulesHandle returns a function which receives HTTP requests and
// publishes the list of routing rules to subscribers.
func (p *Publisher) PublishRoutingRulesHandle() httprouter.Handle {
	return func(respw http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		key := ps.ByName("key")
		rulesID := ps.ByName("rulesID")
		subscriber, ok := p.subscriberIndex[key]
		if !ok {
			// invalid subscriber
			logger.Warnf("Invalid key recieved: %s, response with HTTP 404", key)
			http.NotFound(respw, req)
			return
		}

		for _, allowedRulesID := range subscriber.RoutingRules {
			if allowedRulesID == rulesID {
				goto ALLOWED_RULES
			}
		}

		// not allow
		logger.Warnf("Subscriber(%s) want Routing Rule(%s), but not allow, response with HTTP 404", subscriber.Remarks, rulesID)
		http.NotFound(respw, req)
		return

	ALLOWED_RULES:
		rules, ok := p.routingRulesIndex[rulesID]
		if !ok {
			// config error, should not happen
			logger.Errorf("Subscriber(%s) want Routing Rule(%s), but not found, response with HTTP 500", subscriber.Remarks, rulesID)
			respw.WriteHeader(http.StatusInternalServerError)
			respw.Write([]byte("Server Internal Error, Please tell the administrator."))
			return
		}

		bs, err := jsoniter.ConfigFastest.Marshal(rules)
		if err != nil {
			// marshal error
			logger.Errorf("Marshal routing rule (%s) error: %s", rulesID, err.Error())
			respw.WriteHeader(http.StatusInternalServerError)
			respw.Write([]byte("Server Internal Error, Please tell the administrator."))
			return
		}

		respw.WriteHeader(http.StatusOK)
		respw.Write(bs)
	}
}
