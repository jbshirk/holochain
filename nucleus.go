// Copyright (C) 2013-2017, The MetaCurrency Project (Eric Harris-Braun, Arthur Brock, et. al.)
// Use of this source code is governed by GPLv3 found in the LICENSE file
//----------------------------------------------------------------------------------------
// Nucleus provides an interface for an execution environment interface for chains and their entries
// and factory code for creating nucleii instances

package holochain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NucleusFactory func(h *Holochain, code string) (Nucleus, error)

const (

	// calling types

	STRING_CALLING = "string"
	JSON_CALLING   = "json"

	// exposure types for functions

	// ZOME_EXPOSURE is the default and means the function is only exposed for use by other zomes in the app
	ZOME_EXPOSURE = ""
	// AUTHENTICATED_EXPOSURE means that the function is only available after authentication (TODO)
	AUTHENTICATED_EXPOSURE = "auth"
	// PUBLIC_EXPOSURE means that the function is callable by anyone
	PUBLIC_EXPOSURE = "public"

	// these constants are for a removed feature, see ChangeAppProperty
	// @TODO figure out how to remove code over time that becomes obsolete, i.e. for long-dead changes

	ID_PROPERTY         = "_id"
	AGENT_ID_PROPERTY   = "_agent_id"
	AGENT_NAME_PROPERTY = "_agent_name"
)

var ValidationFailedErr = errors.New("Validation Failed")

// FunctionDef holds the name and calling type of an DNA exposed function
type FunctionDef struct {
	Name        string
	CallingType string
	Exposure    string
}

// ValidExposure verifies that the function can be called in the given context
func (f *FunctionDef) ValidExposure(context string) bool {
	if f.Exposure == PUBLIC_EXPOSURE {
		return true
	}
	return f.Exposure == context
}

// Nucleus type abstracts the functions of code execution environments
type Nucleus interface {
	Type() string
	ValidateAction(action Action, def *EntryDef, pkg *ValidationPackage, sources []string) (err error)
	ValidatePackagingRequest(action ValidatingAction, def *EntryDef) (req PackagingReq, err error)
	ChainGenesis() error
	Call(fn *FunctionDef, params interface{}) (interface{}, error)
	Run(code string) (result interface{}, err error)
}

var nucleusFactories = make(map[string]NucleusFactory)

// RegisterNucleus sets up a Nucleus to be used by the CreateNucleus function
func RegisterNucleus(name string, factory NucleusFactory) {
	if factory == nil {
		panic("Nucleus factory for type %s does not exist." + name)
	}
	_, registered := nucleusFactories[name]
	if registered {
		panic("Nucleus factory for type %s already registered. " + name)
	}
	nucleusFactories[name] = factory
}

// RegisterBultinNucleii adds the built in nucleus types to the factory hash
func RegisterBultinNucleii() {
	RegisterNucleus(ZygoNucleusType, NewZygoNucleus)
	RegisterNucleus(JSNucleusType, NewJSNucleus)
}

// CreateNucleus returns a new Nucleus of the given type
func CreateNucleus(h *Holochain, nucleusType string, code string) (Nucleus, error) {

	factory, ok := nucleusFactories[nucleusType]
	if !ok {
		// Factory has not been registered.
		// Make a list of all available nucleus factories for error.
		var available []string
		for k := range nucleusFactories {
			available = append(available, k)
		}
		sort.Strings(available)
		return nil, fmt.Errorf("Invalid nucleus name. Must be one of: %s", strings.Join(available, ", "))
	}

	return factory(h, code)
}
