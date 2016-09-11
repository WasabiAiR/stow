# Google Cloud Storage Stow Implementation

Location = Google Cloud Storage

Container = Bucket

Item = File

---

Configuration... You need to create a project in google, and then create a service account in google tied to that project. You will need to download a `.json` file with the configuration for the service account. To run the test suite, the service account will need edit privileges inside the project.

To run the test suite, set the `GOOGLE_CREDENTIALS_FILE` environment variable to point to the location of the .json file containing the service account credentials, otherwise the test suite will not be run.

---

Concerns:

- Google's storage plaform is more _eventually consistent_ than other platforms. Sometimes, the tests appear to be flaky because of this. One example is when deleting files from a bucket, then immediately deleting the bucket...sometimes the bucket delete will fail saying that the bucket isn't empty simply because the file delete messages haven't propagated through Google's infrastructure. We may need to add some delay into the test suite to account for this.
