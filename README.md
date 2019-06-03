# Blockchain-based Job Market

video demo: https://youtu.be/7CAsU0Vj_X0

![diagram](https://raw.githubusercontent.com/Simonl07/blockchain-jobmarket/master/diagram.png)

### Three players:
* Applicant
* Employer
* Miners

### Three types of Transaction:
* Merit/Application
* Acceptance
* Acceptance Confirmation

## Sunny day scenario:
* Applicant broadcast TX to miners with TX fees
* Miner put TX into block & finalize TX fees
* Company view Applications on BC
* Company broadcast Acceptance TX to miners with TX fees
* Miner put TX into block & finalize TX fees
* Applicant view BC and respond to Acceptance TX(I accept & encrypted identity) by broadcast a new TX with TX fees
* Miner put TX into block & finalize TX fees
* Company view BC, retrieve Identity, and verify original signatures
* Company contact individual directly to proceed further

## What if a Company received bad Acceptance Confirmation?
Discard

## Other Infrastructures:
* An applicant portal that can: Fill the full application, while the identity is stored on server, View the system from a application perspective: Open, Accepted Companies, Confirm and Release Identity
* A company portal that can: Register the company, View and search all Merits, Accept Merits, View the status of their acceptance.
* A separate infrastructure is available for translating company’s public key to company’s profile, so that applicants can refer to this service to identify the company
