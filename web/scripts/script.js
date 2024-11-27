function get_xml_versions() {
  var xmlHttp = new XMLHttpRequest();
  xmlHttp.open("GET", "http://localhost:8080/available-xml-versions", false);
  xmlHttp.send(null);
  return xmlHttp.responseText;
}

function add_xml_versions_option_elements() {
  var xml_versions = JSON.parse(get_xml_versions());

  var select_element = document.getElementById("version-select");

  xml_versions.forEach((item) => {
    var option = document.createElement("option");
    option.innerText = item;
    option.value = item;
    select_element.appendChild(option);
  });
}

function handleFileSelect(event) {
  const file = event.target.files[0];
  if (file && file.type === "text/xml") {
    processFile(file);
  } else {
    displayResult("Please upload a valid XML file.", "error");
  }
}

function handleFileDrop(event) {
  event.preventDefault();
  const file = event.dataTransfer.files[0];
  if (file && file.type === "text/xml") {
    processFile(file);
  } else {
    displayResult("Please upload a valid XML file.", "error");
  }
}

// Process the selected or dropped XML file
function processFile(file) {
  const version = document.getElementById("version-select").value;
  const formData = new FormData();
  formData.append("file", file);
  formData.append("version", version);

  fetch("/validate-xml", {
    method: "POST",
    body: formData,
  })
    .then((response) => response.json())
    .then((data) => {
      if (data.valid) {
        displayResult("Validation successful: " + data.message, "success");
      } else {
        displayResult("Validation failed: " + data.message, "error");
      }
    })
    .catch((error) => {
      displayResult("An error occurred during validation.", error);
    });
}

// Display validation results
function displayResult(message, status) {
  const resultElement = document.getElementById("result");
  resultElement.textContent = message;
  resultElement.className = status;
}

add_xml_versions_option_elements();
