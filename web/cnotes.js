// JavaScript to respond to a click on Chinese text to show vocabulary dialog.
const corpusTextDiv = document.getElementById('CorpusTextDiv');
const vocabDialog = document.getElementById('vocabDialog');
const okButton = document.getElementById('okButton');
const vocabDialogCn = document.getElementById('vocabDialogCn');
const vocabDialogEn = document.getElementById('vocabDialogEn');

if (corpusTextDiv) {
  corpusTextDiv.addEventListener('click', (evt) => {
    if (typeof vocabDialog.show !== "function") {
    	alert("Cannot show a dialog in this browser");
    	return;
    }
  	console.log(`Got a click from ${evt.target}`);
  	if (!evt.target.innerText || !evt.target.title) {
  		console.log(`Either no text or no equiv for event target ${evt.target}`);
  		return;
  	}
  	if (!vocabDialogCn || !vocabDialogEn) {
  		console.log("Either vocabDialogCn or vocabDialogEn not found");
  		return;
  	}
  	const txt = evt.target.innerText; 
  	const equiv = evt.target.title; 
  	vocabDialogCn.innerText =txt;
  	vocabDialogEn.innerText =equiv;
  	//const x = evt.clientX;
  	//const y = evt.clientY;
  	//console.log(`x: ${x}, y: ${y}`);
    // TO-DO position the dialog in the right place by creating a
    // HTMLDialogElement dynamically
    vocabDialog.show();
  });
}

if (okButton) {
  okButton.addEventListener('click', () => {
    if (typeof vocabDialog.show === "function") {
      vocabDialog.close();
    }
  });
}
