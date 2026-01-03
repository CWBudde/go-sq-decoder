const dropZone = document.getElementById("drop-zone");
const fileInput = document.getElementById("file-input");
const convertButton = document.getElementById("convert-button");
const statusText = document.getElementById("status-text");
const statusDetail = document.getElementById("status-detail");
const fileName = document.getElementById("file-name");
const fileSize = document.getElementById("file-size");
const downloadLink = document.getElementById("download-link");
const logicToggle = document.getElementById("logic-toggle");
const floatToggle = document.getElementById("float-toggle");

let currentFile = null;
let outputUrl = null;
let wasmReady = false;

const formatBytes = (bytes) => {
  if (!bytes && bytes !== 0) return "";
  const units = ["B", "KB", "MB", "GB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex += 1;
  }
  return `${size.toFixed(size < 10 && unitIndex > 0 ? 1 : 0)} ${units[unitIndex]}`;
};

const setStatus = (primary, detail = "") => {
  statusText.textContent = primary;
  statusDetail.textContent = detail;
};

const setFile = (file) => {
  currentFile = file;
  resetOutput();
  if (!file) {
    fileName.textContent = "No file selected";
    fileSize.textContent = "";
    convertButton.disabled = true;
    downloadLink.classList.add("hidden");
    return;
  }
  fileName.textContent = file.name;
  fileSize.textContent = formatBytes(file.size);
  convertButton.disabled = !wasmReady;
  downloadLink.classList.add("hidden");
};

const resetOutput = () => {
  if (outputUrl) {
    URL.revokeObjectURL(outputUrl);
    outputUrl = null;
  }
};

const handleDrop = (event) => {
  event.preventDefault();
  dropZone.classList.remove("is-dragover");
  const file = event.dataTransfer.files[0];
  if (file) {
    setFile(file);
  }
};

const handleDragOver = (event) => {
  event.preventDefault();
  dropZone.classList.add("is-dragover");
};

const handleDragLeave = () => {
  dropZone.classList.remove("is-dragover");
};

const handleFileInput = (event) => {
  const file = event.target.files[0];
  setFile(file);
};

const ensureWasm = async () => {
  if (wasmReady) return;
  if (!window.Go) {
    setStatus("Missing wasm_exec.js", "Copy Go's wasm_exec.js into the web folder.");
    return;
  }
  const go = new window.Go();
  try {
    const response = await fetch("sqdecoder.wasm");
    let result;
    if (WebAssembly.instantiateStreaming) {
      result = await WebAssembly.instantiateStreaming(response, go.importObject);
    } else {
      const bytes = await response.arrayBuffer();
      result = await WebAssembly.instantiate(bytes, go.importObject);
    }
    go.run(result.instance);
    wasmReady = true;
    setStatus("Decoder ready", "Drop a file or choose one to begin.");
    convertButton.disabled = !currentFile;
  } catch (error) {
    setStatus("Failed to load decoder", error.message || String(error));
  }
};

const decodeFile = async () => {
  if (!currentFile || !wasmReady) {
    return;
  }
  resetOutput();
  convertButton.disabled = true;
  setStatus("Decoding...", "Running SQ2 decoder in WebAssembly.");

  try {
    const arrayBuffer = await currentFile.arrayBuffer();
    const inputBytes = new Uint8Array(arrayBuffer);
    const result = window.sqDecodeWav(inputBytes, {
      logic: logicToggle.checked,
      float32: floatToggle.checked,
    });
    if (result && result.error) {
      throw new Error(result.error);
    }
    if (!result || !result.data) {
      throw new Error("Decoder returned no data");
    }
    const outputBytes = result.data;
    const blob = new Blob([outputBytes], { type: "audio/wav" });
    outputUrl = URL.createObjectURL(blob);
    const baseName = currentFile.name.replace(/\.wav$/i, "");
    downloadLink.href = outputUrl;
    downloadLink.download = `${baseName || "decoded"}_quad.wav`;
    downloadLink.classList.remove("hidden");
    setStatus("Done", "Download the decoded WAV file.");
  } catch (error) {
    setStatus("Decode failed", error.message || String(error));
  } finally {
    convertButton.disabled = !currentFile;
  }
};

dropZone.addEventListener("dragover", handleDragOver);
dropZone.addEventListener("dragleave", handleDragLeave);
dropZone.addEventListener("drop", handleDrop);
dropZone.addEventListener("click", () => fileInput.click());
dropZone.addEventListener("keydown", (event) => {
  if (event.key === "Enter" || event.key === " ") {
    event.preventDefault();
    fileInput.click();
  }
});
fileInput.addEventListener("change", handleFileInput);
convertButton.addEventListener("click", decodeFile);

ensureWasm();
