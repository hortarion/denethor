const socket = new WebSocket(`ws://${window.location.hostname}:8080/ws`);

socket.onopen = () => {
  console.log("Connected to backend server");
  const message = {
    token: "sys",
    data: "Client connected",
  };
  socket.send(JSON.stringify(message));
};

socket.onmessage = (event) => {
  // DEBUG LOG
  console.log("Message from server:", event.data);
  handleInputUpdate(event.data);
};

socket.onclose = () => {
  console.log("Disconnected from backend server");
};

const output = document.getElementById("output");
const inputField = document.getElementById("inputField");
const submitButton = document.getElementById("submitButton");
const themeToggle = document.getElementById("themeToggle");
const body = document.body;
const MAX_LINES = 50;

let outputLines = [];
let maskedInput = false;

function updateOutputLDisplay() {
  output.textContent = outputLines.join("\n");
  output.scrollTop = output.scrollHeight;
}

function appendToOutput(text) {
  if (maskedInput) {
    text = text.split(">> ")[1];
    var maskedText = "";
    for (i = 0; i < text.length; i++) {
      maskedText += "*";
    }
    const message = {
      token: "auth",
      data: text,
    };
    text = ">> " + maskedText;
    socket.send(JSON.stringify(message));
  }
  outputLines.push(text);
  if (outputLines.length > MAX_LINES) {
    outputLines = outputLines.slice(-MAX_LINES);
  }
  updateOutputLDisplay();
}

submitButton.addEventListener("click", () => {
  const inputText = inputField.value;
  if (inputText.length === 0) {
    return;
  }
  if (inputText.trim() !== "") {
    appendToOutput(">> " + inputText);
    inputField.value = "";
  }
  output.dispatchEvent(new Event("outputUpdated"));
});

inputField.addEventListener("keypress", (e) => {
  if (e.key === "Enter") {
    submitButton.click();
  }
});

themeToggle.addEventListener("click", () => {
  body.classList.toggle("light-mode");
});

window.printToOutput = function (text) {
  appendToOutput("<< " + text);
  output.dispatchEvent(new Event("outputUpdated"));
};

window.clearOutput = function () {
  outputLines = [];
  updateOutputLDisplay();
};

function printToPage(text) {
  window.printToOutput(text);
}

async function handleOutputUpdate() {
  const lastLine = outputLines[outputLines.length - 1];
  if (lastLine.startsWith(">>")) {
    const text = lastLine.split(">> ")[1];

    const message = {
      token: "console",
      data: text,
    };
    socket.send(JSON.stringify(message));
  }
}

function handleInputUpdate(input) {
  const message = JSON.parse(input);
  if (message.token === "auth" && message.data === "type in your password") {
    if (maskedInput === false) {
      maskedInput = true;
      inputField.type = "password";
      inputField.placeholder = "Enter password here";
      return;
    }
  }
  if (maskedInput === true) {
    maskedInput = false;
    inputField.type = "text";
    inputField.placeholder = "Enter text here";
    return;
  }

  if (message.data.length == 0) {
    return;
  }
  if (message.token === "console") {
    if (message.data === "clear") {
      clearOutput();
    }
  }
  printToPage(message.data);
}

// Event listener
output.addEventListener("outputUpdated", async () => {
  await handleOutputUpdate();
});

// Ensure the DOM is fully loaded before attaching listeners
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => {
    // Call printToPage when the DOM is ready
    console.log("DOM listening for events");
  });
} else {
  // If the DOM is already loaded, call printToPage immediately
  console.log("DOM already listening");
}

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
