// Respond to a click on Chinese text to show vocabulary dialog.
const corpusTextDiv = document.getElementById('CorpusTextDiv');

if (corpusTextDiv) {
  corpusTextDiv.addEventListener('click', (evt) => {
  	console.log(`Got a click from ${evt.target}`);
    if (evt == null || evt.target == null) {
      console.log('evt is null');
      return;
    }
    const elem = evt.target as HTMLElement;
  	if (!elem.innerText || !elem.title) {
  		console.log(`Either no text or no equiv for event target ${evt.target}`);
  		return;
  	}
    let dialog = document.getElementById('dialog');
    if (!dialog) {
      dialog = document.createElement('dialog');
      dialog.id = 'dialog';
      if (elem.parentElement == null) {
        console.log('elem.parentElement is null');
        return;
      }
      elem.parentElement.appendChild(dialog);
    }
    const diag = dialog as HTMLDialogElement;
    if (typeof diag.show !== 'function') {
      alert('Cannot show a dialog in this browser');
      return;
    }
    let vocabDialogCn = document.getElementById('vocabDialogCn');
    if (!vocabDialogCn) {
      vocabDialogCn = document.createElement('p');
      vocabDialogCn.id = 'vocabDialogCn';
      vocabDialogCn.innerText = elem.innerText;
      diag.appendChild(vocabDialogCn);
    }
    vocabDialogCn.innerText = elem.innerText;

    let vocabDialogEn = document.getElementById('vocabDialogEn');
    if (!vocabDialogEn) {
      vocabDialogEn = document.createElement('p');
      vocabDialogEn.id = 'vocabDialogEn';
      diag.appendChild(vocabDialogEn);
    }
    vocabDialogEn.innerText = elem.title;

    let okButton = document.getElementById('okButton');
    if (!okButton) {
      okButton = document.createElement('button');
      okButton.id = 'okButton';
      okButton.innerText = 'OK';
      diag.appendChild(okButton);
      okButton.addEventListener('click', () => {
        diag.close();
      });
    }
    diag.style.position = 'relative;';
    console.log(`X: ${evt.clientX}, Y:  ${evt.clientY}`);
    diag.style.left = '20px;';
    diag.style.top = '20px;';
    diag.show();
  });
}
