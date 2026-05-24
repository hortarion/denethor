function setupWebSocket() {
  const jwt = localStorage.getItem("jwt");
  const socket = new WebSocket(
    `ws://${window.location.hostname}:8080/ws?jwt=${jwt}`,
  );

  socket.onopen = () => {
    console.log("Connected to backend server");

    /* if (jwt && jwt !== "authenticated") {
      socket.send(
        JSON.stringify({
          channel: "auth",
          token: "jwt",
          data: jwt,
        }),
      );
      } */
    const message = {
      channel: "sys",
      token: "",
      data: "Client connected",
    };
    socket.send(JSON.stringify(message));
  };
  return socket;
}

let socket = setupWebSocket();

function reconnectWebSocket() {
  if (socket) {
    socket.close();
  }
  socket = setupWebSocket();
}

function storeJWT(token) {
  localStorage.setItem("jwt", token);
}

function clearJWT() {
  console.log("Clearing jwt");
  localStorage.removeItem("jwt");
  userField.textContent = "Guest";
}

socket.onmessage = (event) => {
  const packet = JSON.parse(event.data);
  if (packet.channel === "sys" && packet.token === "jwt") {
    localStorage.setItem("jwt", packet.token);
  }
  // DEBUG LOG
  console.log("Packet from server:", packet);
  handleInputUpdate(packet);
};

socket.onclose = () => {
  console.log("Disconnected from backend server");
};

const userField = document.getElementById("userField");
const output = document.getElementById("output");
const inputField = document.getElementById("inputField");
const submitButton = document.getElementById("submitButton");
const themeToggle = document.getElementById("themeToggle");
const body = document.body;
const MAX_LINES = 50;

let outputLines = [];
let maskedInput = false;

submitButton.addEventListener("click", () => {
  const inputText = inputField.value;
  if (inputText.length === 0) {
    return;
  }
  if (inputText.trim() !== "") {
    const now = new Date();
    const timestamp = `${now.getHours()}:${now.getMinutes()}:${now.getSeconds()}`;
    appendToOutput(`[${timestamp}] You: ` + inputText);
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
  if (text.length === 0) {
    return;
  }
  const now = new Date();
  const timestamp = `${now.getHours()}:${now.getMinutes()}:${now.getSeconds()}`;
  appendToOutput(`[${timestamp}] Denethor:\n` + text);
  output.dispatchEvent(new Event("outputUpdated"));
};

window.clearOutput = function () {
  outputLines = [];
  updateOutputDisplay();
};

function updateOutputDisplay() {
  output.textContent = outputLines.join("\n");
  output.scrollTop = output.scrollHeight;
}

function printToPage(text) {
  window.printToOutput(text);
}

function appendToOutput(text) {
  const denethorRegex = /\[.*?\] Denethor:/;
  if (denethorRegex.test(text)) {
    outputLines.push(text);
    if (outputLines.length > MAX_LINES) {
      outputLines = outputLines.slice(-MAX_LINES);
    }
    updateOutputDisplay();
    return;
  }
  if (maskedInput) {
    const now = new Date();
    const timestamp = `${now.getHours()}:${now.getMinutes()}:${now.getSeconds()}`;
    const match = text.match(/\[.*?\] You: (.*)/);
    const command = match ? match[1] : text;
    var maskedText = "*".repeat(command.length);
    const message = {
      channel: "auth",
      token: "",
      data: command,
    };
    text = `[${timestamp}] You: ${maskedText}`;
    socket.send(JSON.stringify(message));
  }
  outputLines.push(text);
  if (outputLines.length > MAX_LINES) {
    outputLines = outputLines.slice(-MAX_LINES);
  }
  updateOutputDisplay();
}

async function handleOutputUpdate() {
  const lastLine = outputLines[outputLines.length - 1];
  const regexDenethor = /\[.*?\] Denethor: /;
  if (regexDenethor.test(lastLine)) {
    return;
  }
  const regex = /\[.*?\] You: /;
  if (regex.test(lastLine)) {
    const commands = lastLine.split(/\[.*?\] You: /)[1];
    const message = {
      channel: "console",
      token: "",
      data: commands,
    };
    socket.send(JSON.stringify(message));
  }
}

function handleInputUpdate(packet) {
  switch (packet.channel) {
    case "auth":
      if (packet.token === "password") {
        maskedInput = true;
        inputField.type = "password";
        inputField.placeholder = "Enter your password";
        return;
      } else {
        maskedInput = false;
        inputField.type = "text";
        inputField.placeholder = "Enter text here";
        printToPage(packet.data);
        return;
      }
    case "console":
      if (packet.data.length === 0) {
        return;
      }
      printToPage(packet.data);
      break;
    case "sys":
      handleSystemMessage(packet);
      break;
    default:
      console.log("Unknown input package", packet);
  }
}

function handleSystemMessage(message) {
  switch (message.token) {
    case "clear":
      clearOutput();
      break;
    case "authenticated":
      userField.textContent = "Logged in as " + message.data;
      break;
    case "logout":
      clearJWT();
      maskedInput = false;
      inputField.type = "text";
      inputField.placeholder = "Enter text here";
      break;
    case "JWT":
      storeJWT(message.data);
      break;
    default:
      console.log("Unknown system message token:", message.token);
  }
}

output.addEventListener("outputUpdated", async () => {
  await handleOutputUpdate();
});

output.addEventListener("visibilitychange", () => {
  if (document.visibilityState === "visible") {
    reconnectWebSocket();
  }
});

window.addEventListener("online", reconnectWebSocket);

window.addEventListener("beforeunload", () => {
  if (socket) {
    socket.close();
  }
});

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => {
    console.log("DOM listening for events");
  });
} else {
  console.log("DOM already listening");
}

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
