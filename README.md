# Blockchain-based Job Market

https://youtu.be/7CAsU0Vj_X0 demo

![diagram](https://raw.githubusercontent.com/usfcs686/cs686-blockchain-p3-Simonl07/master/diagram.png?token=AFC2RFW6C5F5ZCH6LHE5BLS4436B4)

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


## Final Progress Report:

* Full process/Sunny day senario completes
* Added applicant command line interface to generate keys, submit applications, confirm acceptances, and view their merits status
* Added employer command line interface to generate keys, accept merits, and verify acceptance confirmation
* Added a "producer" field to block so we know which miner created this block
* Did not add a separate in memory data structure for every miner to keep track of each miner's balance, instead, an api endpoint is added to the miner to display all balances by traversing though the canonical chain
* Transaction verifications: Merits can always be published, acceptance need to be accepting a valid merit, confirmation needs to be confiming on a valid acceptance
* Added optional canonical look back capabilities so that after n blocks back(after the first height with no forks) are considered canonical
* Added two types of assymetric keys: ECDSA & RSA, ECDSA are used for transaction signatures due to smaller key size and signature size. RSA is used for encrypting the identity of applicants
* Miner's transaction pool is essentially a priority queue based on transaction fees
* Did not have time for creating a separate infrastructure for transalating employer public keys to employer profile. The opportunity cost is a bit high for the completion of the project. Practically, companies could post their public key on the company website for applicant's reference.
* Rainy day senarios:
  * What if a merits receives many acceptances? > Does not matter, it is still applicant's choice to determine which to confirm
  * What if an acceptance receives multiple confirmations? > Does not matter, only the confirmations with correct identity is valid
  
  
### Some potential issues:
1. I realized half way through the implementation that I am merging the infrastructure layer(the block chain transactions) with the application layer a bit in the way that I am using merit hashes(which is an application concept) as the recipient of acceptance. There are few other cases that involves this blending. This is probably a bad idea but it requires lots of changes to refactor and will make my payload structure a bit more complicated (for example right now the acceptance type transaction's payload is just the RSA public key of employer, the target merit is stored in Transaction.To. Ideally I should have it inside the payload too and the Transaction.To should be refering to the transaction that holds the target merit)
2. The role of validator and miner is merged in this ecosystem. Miner is both responsible for the functioning and reliability of the system (transaction & block verification, serving expensive apis for front end, and storing the entire block chain) and use as many computation power as possible to create blocks. These two goals are conflicting with each other especially given that my current implementation for validating is not the most efficient way (traversing though entire history of transactions). This may be a problem.
3. Ideally all command line interfaces in the client folder should be aggregated in a centralized website that provides those services (generate and store keys, view status, submit application, confirm acceptance and release identity), and also an employer dashboard (browse[maybe a search engine that seach keywords in merits], accept a merit, and verify identity). This requires lots of extra work and I also do not have lots of experience writing front end pages, however, if there is this website, the user experience will be much better and coherent. The current model also lacks a notification system which can be made possible with a website serivce.
