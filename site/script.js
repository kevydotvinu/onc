document.getElementById('hostPrefix').value = 23;
document.getElementById('clusterNetwork').value = '10.128.0.0/14';
document.getElementById('serviceNetwork').value = '172.30.0.0/16';
document.getElementById('machineNetwork').value = '192.168.1.0/24';

function isValidCIDR(input) {
    const cidrPattern = /^(?:\d{1,3}\.){3}\d{1,3}\/(1[0-9]|2[0-9]|3[0-2]|[1-9])$/;
    return cidrPattern.test(input);
}

function isValidHostPrefix(hostPrefix) {
    const parsedHostPrefix = parseInt(hostPrefix);
    return !isNaN(parsedHostPrefix) && parsedHostPrefix >= 1 && parsedHostPrefix <= 32;
}

function calculateNetwork() {
    const hostPrefix = document.getElementById('hostPrefix').value;
    const clusterNetwork = document.getElementById('clusterNetwork').value;
    const serviceNetwork = document.getElementById('serviceNetwork').value;
    const machineNetwork = document.getElementById('machineNetwork').value;
    const cni = document.getElementById('cni').value;

    // Simple validation
    if (!isValidHostPrefix(hostPrefix) || !isValidCIDR(clusterNetwork) || !isValidCIDR(serviceNetwork) || !isValidCIDR(machineNetwork)) {
        alert('Please fill in all fields correctly. Network e.g, 192.168.1.0/24 and HostPrefix from 1 to 32).');
        return;
    }

    // Prepare the request object
    const request = {
        hostPrefix: parseInt(hostPrefix),
        clusterNetwork: clusterNetwork,
        serviceNetwork: serviceNetwork,
        machineNetwork: machineNetwork,
	cni: cni
    };

    // Send the request to the Go server
    const host = location.hostname === "localhost" ? "https://onc.netlify.app" : "";
    const url = host + "/.netlify/functions/onc"
    fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify(request),
    })
        .then(response => response.json())
        .then(data => {
            // Display the response
            document.getElementById('result').innerText = JSON.stringify(data, null, 2);
        })
        .catch(error => {
            alert('Error sending request to server: ' + error);
        });
}
