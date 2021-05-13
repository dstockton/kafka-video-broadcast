// https://jsfiddle.net/43rm7258/1/

const constraints = {audio: true, video: {width: 426, height: 240}};
const options = { mimeType: "video/webm;codecs=opus, vp9" };

const remoteVideo = document.getElementById('remote_video');
const localVideo = document.getElementById('local_video');

const mediaSource = new MediaSource();

let sessionID = (new Date().getTime()) + '-' + Math.random()*100000000000000000;
let ws = new WebSocket((window.location.protocol === "https:" ? "wss://" : "ws://") + window.location.host + '/video/connections?sessionID=' + sessionID);

let receivedBlobs = [];
let sourceBuffer;
let updatingBuffer = false;

navigator.mediaDevices.getUserMedia(constraints).then(function(userMediaStream) {
  localVideo.srcObject = userMediaStream;  

  let mediaRecorder = new MediaRecorder(userMediaStream, options);
  mediaRecorder.ondataavailable = handleDataAvailable;
  mediaRecorder.start(100);
});

mediaSource.addEventListener("sourceopen", function(){
  console.log("Adding source buffer to media source...");
  sourceBuffer = mediaSource.addSourceBuffer(options.mimeType);

  sourceBuffer.addEventListener("updateend", () => {
    updatingBuffer = false;
  });
});
remoteVideo.src = window.URL.createObjectURL(mediaSource);

function handleDataAvailable(event) {
  var blobby = event.data;
  if (blobby && blobby.size > 0) {
    while (blobby.size > 32000) {
      console.log("Splitting large frame! ", blobby.size);
      ws.send(blobby.slice(0,32000));
      blobby = blobby.slice(32000);
    }
    ws.send(blobby);
  }

}

ws.onmessage = (evt) => {
  // Store the received blob (we can't always process it immediately)
  receivedBlobs.push(evt.data);

  if (receivedBlobs.length > 5) {
    if (receivedBlobs.length === 5)
      console.log("buffered enough for delayed playback");
    if (!updatingBuffer) {
      updatingBuffer = true;
      const receivedBlob = receivedBlobs.shift();
      receivedBlob.arrayBuffer().then(function(arrBuff) {
        if (!sourceBuffer.updating) {
          sourceBuffer.appendBuffer(arrBuff);
        } else {
          console.log("Media source buffer is busy - re-queueing...")
          receivedBlobs.unshift(receivedBlob);
        }
      });
    }
  } else {
    console.log("Still buffering from websocket...")
  }
};
