auth to google:
  - go to your project in google console and navigate to the credentials pane https://console.cloud.google.com/apis/credentials. 
  - click Create Credentials - Service Account
    - give it a unique name and hit continue
    - add a role: Cloud Storage - Admin
    - hit Done
  - click on the newly created service account and go to Keys tab
    - add a key - Create new key
    - JSON
  - (OPTIONAL: download the json key and encode it with base64, eg: cat key.json | base64)
  - add the key to your .env and cloud env

creating a bucket:
  - name must be unique
  - location type: Region
  - location eg: us-central1 (Iowa)
  - default storage class: Standard (for images), Nearline (for backups), ...
  - access control: Uniform (bucket level permissions)

making the bucket publically available:
  - go to https://console.cloud.google.com/storage/browser
  - click on your bucket
  - permissions tab
  - click on Grant Access
  - New principal: allUsers
  - Role: Storage Object Viewer

in order to be able to upsert the files to a bucket we need not only create role to that bucket but also read&delete... basically Admin role