//
// Copyright 2015, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package multisyncer

type syncer chan cmdData

type cmdData struct {
	action cmdAction
	token  interface{}
	done   chan struct{}
}

type cmdAction int

const (
	lock cmdAction = iota
	unlock
)

// MultiSyncer synchronizes access based on a given token
type MultiSyncer interface {
	Lock(interface{})
	Unlock(interface{})
}

// New returns a new MultiSyncer
func New() MultiSyncer {
	s := make(syncer)
	go s.run()
	return s
}

func (s syncer) run() {
	store := make(map[interface{}]chan struct{})

	for cmd := range s {
		go func(cmd cmdData) {
			switch cmd.action {
			case lock:
				if c, ok := store[cmd.token]; ok {
					c <- struct{}{}
				} else {
					c := make(chan struct{}, 1)
					c <- struct{}{}
					store[cmd.token] = c
				}
			case unlock:
				if c, ok := store[cmd.token]; ok {
					<-c
				}
			}

			cmd.done <- struct{}{}
		}(cmd)
	}
}

// Lock implements the MultiSyncer interface
func (s syncer) Lock(token interface{}) {
	done := make(chan struct{})
	s <- cmdData{action: lock, token: token, done: done}
	<-done
}

// Unlock implements the MultiSyncer interface
func (s syncer) Unlock(token interface{}) {
	done := make(chan struct{})
	s <- cmdData{action: unlock, token: token, done: done}
	<-done
}
