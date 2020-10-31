// JavaScript to respond to a click on Chinese text to show vocabulary dialog.
const corpusTextDiv = document.getElementById('CorpusTextDiv');

if (corpusTextDiv) {
  corpusTextDiv.addEventListener('click', (evt) => {
  	console.log(`Got a click from ${evt.target}`);
  	if (!evt.target.innerText || !evt.target.title) {
  		console.log(`Either no text or no equiv for event target ${evt.target}`);
  		return;
  	}
    let dialog = document.getElementById('dialog');
    if (!dialog) {
      dialog = document.createElement('dialog');
      dialog.id = 'dialog';
      evt.target.parentElement.appendChild(dialog);
    }
    if (typeof dialog.show !== 'function') {
      alert('Cannot show a dialog in this browser');
      return;
    }
    let vocabDialogCn = document.getElementById('vocabDialogCn');
    if (!vocabDialogCn) {
      vocabDialogCn = document.createElement('p');
      vocabDialogCn.id = 'vocabDialogCn';
      vocabDialogCn.innerText = evt.target.innerText;
      dialog.appendChild(vocabDialogCn);
    }
    vocabDialogCn.innerText = evt.target.innerText;

    let vocabDialogEn = document.getElementById('vocabDialogEn');
    if (!vocabDialogEn) {
      vocabDialogEn = document.createElement('p');
      vocabDialogEn.id = 'vocabDialogEn';
      dialog.appendChild(vocabDialogEn);
    }
    vocabDialogEn.innerText = evt.target.title;

    let okButton = document.getElementById('okButton');
    if (!okButton) {
      okButton = document.createElement('button');
      okButton.id = 'okButton';
      okButton.innerText = 'OK';
      dialog.appendChild(okButton);
      okButton.addEventListener('click', () => {
        dialog.close();
      });
    }
    dialog.style.position = 'relative;';
    console.log(`X: ${evt.clientX}, Y:  ${evt.clientY}`);
    dialog.style.left = '20px;';
    dialog.style.top = '20px;';
    dialog.show();
  });
}
