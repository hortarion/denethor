/*
  EMBEDD INTO INDEX.HTML?
*/

// Extend the Window interface to include printToOutput
declare global {
  interface Window {
    printToOutput: (text: string) => void;
  }
}

const output = document.getElementById("output") as HTMLElement;
const inputField = document.getElementById("inputField") as HTMLTextAreaElement;
const submitButton = document.getElementById("submitButton") as HTMLElement;
const themeToggle = document.getElementById("themeToggle") as HTMLElement;
const body = document.body;

submitButton.addEventListener("click", () => {
  const inputText = inputField.value;
  if (inputText.trim() !== "") {
    output.textContent += ">> " + inputText + "\n";
    inputField.value = "";
    output.scrollTop = output.scrollHeight;
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

window.printToOutput = function (text: string) {
  output.textContent += "<< " + text + "\n";
  output.scrollTop = output.scrollHeight;
  output.dispatchEvent(new Event("outputUpdated"));
};

// Exported function to print a message to the output
export function printToPage(text: string): void {
  window.printToOutput(text);
}

async function handleOutputUpdate(): Promise<void> {
  const lines: string[] = output.textContent.trim().split("\n");
  const lastLine: string = lines[lines.length - 1].trim();
  const input: string = lastLine.split(" ")[1];

  // API call to backend
  if (lastLine.startsWith(">>")) {
    // console.log(input); // DEBUG log
    try {
      const response = await fetch(`/api/handleInput`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ input }),
      });
      const reply = await response.text();
      if (reply.length > 0) {
        printToPage(reply);
      }
    } catch (err) {
      printToPage(`API error: ${err}`);
    }
  }
}

// Add the event listener for the custom event
output.addEventListener("outputUpdated", async () => {
  console.log("outputUpdated event fired");
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

async function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}
