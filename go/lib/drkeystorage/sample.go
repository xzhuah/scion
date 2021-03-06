// Copyright 2019 ETH Zurich, Anapaya Systems
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package drkeystorage

const drkeyLvl1DBSample = `
# The type of drkeydb backend. (default sqlite)
Backend = "sqlite"

# Connection for the drkey database.
Connection = "/var/lib/scion/drkeydb/%s.drkey.db"

# The maximum number of open connections to the database. In case of the
# empty string, the limit is not set and uses the go default. (default "")
MaxOpenConns = ""

# The maximum number of idle connections to the database. In case of the
# empty string, the limit is not set and uses the go default. (default "")
MaxIdleConns = ""
`

const drkeyDelegationListSample = `
# The list of hosts authorized to get a DS per protocol.
piskes = [ "127.0.0.1", "127.0.0.2"]
`
