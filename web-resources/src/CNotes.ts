/*
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

/**
 *  @fileoverview  Entry point for the dictionary browser app
 */

import { fromEvent } from "rxjs";

import { MDCDialog } from "@material/dialog";
import { MDCDrawer } from "@material/drawer";
import { MDCList } from "@material/list";
import { MDCTopAppBar } from "@material/top-app-bar";

/**
 * A browser app that implements the Chinese-English dictionary web view.
 */
export class CNotes {
  private dialogDiv: HTMLElement;
  private wordDialog: MDCDialog;

  /**
   * @constructor
   */
  constructor() {
    const dialogDiv = document.querySelector("#CnotesVocabDialog");
    if (dialogDiv && dialogDiv instanceof HTMLElement) {
      this.dialogDiv = dialogDiv;
      this.wordDialog = new MDCDialog(dialogDiv);
    } else {
      console.log("Missing #CnotesVocabDialog from DOM");
      const dialogContainer = document.createElement("div");
      dialogContainer.className = "mdc-dialog__container";
      this.dialogDiv = document.createElement("div");
      this.dialogDiv.className = "mdc-dialog";
      this.dialogDiv.appendChild(dialogContainer);
      this.wordDialog = new MDCDialog(this.dialogDiv);
    }
  }

  /**
   * View setup is here
   */
  public init() {
    console.log("CNotes.init");
    this.initDialog();
  }

  /**
   * Shows the vocabular dialog with details of the given word
   * @param {HTMLElement} elem - the element to display the dialog for
   * @param {string} chineseText - text of the headword to display. If not
   *                 provided, the text from the element will be used.
   */
  public showVocabDialog(elem: HTMLElement, chineseText = "") {
    // Show Chinese, pinyin, and English
    const titleElem = this.querySelectorOrNull("#VocabDialogTitle");
    const s = elem.title;
    const n = s.indexOf("|");
    const pinyin = s.substring(0, n);
    let english = "";
    if (n < s.length) {
      english = s.substring(n + 1, s.length);
    }
    let chinese = this.getTextNonNull(elem);
    if (chineseText !== "") {
      chinese = chineseText;
    }
    console.log(`Value: ${chinese}`);
    const pinyinSpan = this.querySelectorOrNull("#PinyinSpan");
    const englishSpan = this.querySelectorOrNull("#EnglishSpan");
    if (titleElem) {
      titleElem.innerHTML = chinese;
    }
    if (pinyinSpan) {
      pinyinSpan.innerHTML = pinyin;
    }
    if (englishSpan) {
      if (english) {
        englishSpan.innerHTML = english;
      } else {
        englishSpan.innerHTML = "";
      }
    }
    this.wordDialog.open();
  }

  // Gets DOM element text content checking for null
  private getTextNonNull(elem: HTMLElement): string {
    const chinese = elem.textContent;
    if (chinese === null) {
      return "";
    }
    return chinese;
  }

  /** Initialize dialog so that it can be shown when user clicks on a Chinese
   *  word.
   */
  private initDialog() {
    const dialogDiv = document.querySelector("#CnotesVocabDialog");
    if (!dialogDiv) {
      console.log("initDialog no dialogDiv");
      return;
    }
    const clicks = fromEvent(document, "click");
    clicks.subscribe((e) => {
      if (e.target && e.target instanceof HTMLElement) {
        const t = e.target as HTMLElement;
        if (t.matches(".vocabulary")) {
          e.preventDefault();
          this.showVocabDialog(t);
          return false;
        }
      }
    });
  }

  // Looks up an element checking for null
  private querySelectorOrNull(selector: string): HTMLElement | null {
    const elem = document.querySelector(selector);
    if (elem === null) {
      console.log(`Unexpected missing HTML element ${ selector }`);
      return null;
    }
    return elem as HTMLElement;
  }  
}
