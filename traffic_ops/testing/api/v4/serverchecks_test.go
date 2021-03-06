package v4

/*

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/apache/trafficcontrol/lib/go-tc"
	"github.com/apache/trafficcontrol/lib/go-util"
)

func TestServerChecks(t *testing.T) {
	WithObjs(t, []TCObj{CDNs, Types, Parameters, Profiles, Statuses, Divisions, Regions, PhysLocations, CacheGroups, Servers, ServerCheckExtensions, ServerChecks}, func() {
		CreateTestInvalidServerChecks(t)
		UpdateTestServerChecks(t)
		GetTestServerChecks(t)
		GetTestServerChecksWithName(t)
		GetTestServerChecksWithID(t)
	})
}

func CreateTestServerChecks(t *testing.T) {
	SwitchSession(toReqTimeout, Config.TrafficOps.URL, Config.TrafficOps.Users.Admin, Config.TrafficOps.UserPassword, Config.TrafficOps.Users.Extension, Config.TrafficOps.UserPassword)

	for _, servercheck := range testData.Serverchecks {
		resp, _, err := TOSession.InsertServerCheckStatus(servercheck)
		t.Logf("Response: %v host_name %v check %v", *servercheck.HostName, *servercheck.Name, resp)
		if err != nil {
			t.Errorf("could not CREATE servercheck: %v", err)
		}
	}
	SwitchSession(toReqTimeout, Config.TrafficOps.URL, Config.TrafficOps.Users.Extension, Config.TrafficOps.UserPassword, Config.TrafficOps.Users.Admin, Config.TrafficOps.UserPassword)
}

func CreateTestInvalidServerChecks(t *testing.T) {
	toReqTimeout := time.Second * time.Duration(Config.Default.Session.TimeoutInSecs)

	_, _, err := TOSession.InsertServerCheckStatus(testData.Serverchecks[0])
	if err == nil {
		t.Error("expected to receive error with non extension user")
	}

	SwitchSession(toReqTimeout, Config.TrafficOps.URL, Config.TrafficOps.Users.Admin, Config.TrafficOps.UserPassword, Config.TrafficOps.Users.Extension, Config.TrafficOps.UserPassword)

	invalidServerCheck := tc.ServercheckRequestNullable{
		Name:     util.StrPtr("BOGUS"),
		Value:    util.IntPtr(1),
		ID:       util.IntPtr(-1),
		HostName: util.StrPtr("bogus_hostname"),
	}

	// Attempt to create a ServerCheck with invalid server ID
	_, _, err = TOSession.InsertServerCheckStatus(invalidServerCheck)
	if err == nil {
		t.Error("expected to receive error with invalid id")
	}

	invalidServerCheck.ID = nil
	// Attempt to create a ServerCheck with invalid host name
	_, _, err = TOSession.InsertServerCheckStatus(invalidServerCheck)
	if err == nil {
		t.Error("expected to receive error with invalid host name")
	}

	// get valid name to get past host check
	invalidServerCheck.Name = testData.Serverchecks[0].Name

	// Attempt to create a ServerCheck with invalid servercheck name
	_, _, err = TOSession.InsertServerCheckStatus(invalidServerCheck)
	if err == nil {
		t.Error("expected to receive error with invalid servercheck name")
	}
	SwitchSession(toReqTimeout, Config.TrafficOps.URL, Config.TrafficOps.Users.Extension, Config.TrafficOps.UserPassword, Config.TrafficOps.Users.Admin, Config.TrafficOps.UserPassword)
}

func UpdateTestServerChecks(t *testing.T) {
	SwitchSession(toReqTimeout, Config.TrafficOps.URL, Config.TrafficOps.Users.Admin, Config.TrafficOps.UserPassword, Config.TrafficOps.Users.Extension, Config.TrafficOps.UserPassword)
	for _, servercheck := range testData.Serverchecks {
		*servercheck.Value--
		resp, _, err := TOSession.InsertServerCheckStatus(servercheck)
		t.Logf("Response: %v host_name %v check %v", *servercheck.HostName, *servercheck.Name, resp)
		if err != nil {
			t.Errorf("could not update servercheck: %v", err)
		}
	}
	SwitchSession(toReqTimeout, Config.TrafficOps.URL, Config.TrafficOps.Users.Extension, Config.TrafficOps.UserPassword, Config.TrafficOps.Users.Admin, Config.TrafficOps.UserPassword)
}

func GetTestServerChecks(t *testing.T) {
	hostname := testData.Serverchecks[0].HostName
	// Get server checks
	serverChecksResp, alerts, _, err := TOSession.GetServersChecks(nil, nil)
	if err != nil {
		t.Fatalf("could not GET serverchecks: %v (alerts: %+v)", err, alerts)
	}
	found := false
	for _, sc := range serverChecksResp {
		if sc.HostName == *hostname {
			found = true

			if sc.Checks == nil {
				t.Errorf("server %s had no checks - expected it to have at least two", *hostname)
				break
			}

			if ort, ok := sc.Checks["ORT"]; !ok {
				t.Error("no 'ORT' servercheck exists - expected it to exist")
			} else if ort == nil {
				t.Error("'null' returned for ORT value servercheck - expected pointer to 12")
			} else if *ort != 12 {
				t.Errorf("%v returned for ORT value servercheck - expected 12", *ort)
			}

			if ilo, ok := sc.Checks["ILO"]; !ok {
				t.Error("no 'ILO' servercheck exists - expected it to exist")
			} else if ilo == nil {
				t.Error("'null' returned for ILO value servercheck - expected pointer to 0")
			} else if *ilo != 0 {
				t.Errorf("%v returned for ILO value servercheck - expected 0", *ilo)
			}
			break
		}
	}
	if !found {
		t.Errorf("expected to find servercheck for host %v", hostname)
	}
}

func GetTestServerChecksWithName(t *testing.T) {
	params := url.Values{}
	params.Set("hostName", "atlanta-edge-01")

	// Get server checks
	scResp, alerts, _, err := TOSession.GetServersChecks(params, nil)
	if len(scResp) == 0 {
		t.Fatal("no server checks in response, quitting")
	}
	if err != nil {
		t.Fatalf("could not GET serverchecks by name (%v): %v (alerts: %+v)", scResp[0].HostName, err, alerts)
	}
}

func GetTestServerChecksWithID(t *testing.T) {
	params := url.Values{}
	serverChecksResp, _, _, err := TOSession.GetServersChecks(nil, nil)
	if len(serverChecksResp) == 0 {
		t.Fatal("no server checks in response, quitting")
	}
	if serverChecksResp[0].ID == 0 {
		t.Fatal("ID of the response server is nil, quitting")
	}
	id := serverChecksResp[0].ID
	params.Set("id", strconv.Itoa(id))

	// Get server checks
	scResp, alerts, _, err := TOSession.GetServersChecks(params, nil)
	if len(scResp) == 0 {
		t.Fatal("no server checks in response, quitting")
	}
	if err != nil {
		t.Fatalf("could not GET serverchecks by id (%v): %v (alerts: %+v)", scResp[0].ID, err, alerts)
	}
}

// Need to define no-op function as TCObj interface expects a delete function
// There is no delete path for serverchecks
func DeleteTestServerChecks(t *testing.T) {
	return
}
