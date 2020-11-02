/**
 * Licensed  under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import { CNotes } from "./CNotes";
import { CNotesMenu } from "./CNotesMenu";

declare const __VERSION__: string;

/**
 * Entry point for all pages.
 */
console.log(`App version: ${ __VERSION__ }, online: ${ navigator.onLine }`);
const menu = new CNotesMenu();
menu.init();
// Initialize the dictionary
const app = new CNotes();
app.init();
