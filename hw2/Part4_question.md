In this test, I launched two EC2 instances, each running the same album-server Docker container on port 8080. Initially, both instances returned the same default list of albums.

Then, I sent a POST request with a new album entry ("The Modern Sound of Betty Carter") to one of the instances (ec2-44-242-201-88). Afterward, I re-sent GET requests to both instances.

**Observation**: Only the instance that received the POST showed the new album; the other instance still displayed the original list of three albums.

**Conclusion**: This confirms that the two EC2 instances are completely independent. Since they each run their own local in-memory version of the application and do not share a database or any persistent backend, the changes made to one are not reflected in the other. This is a common situation in stateless Docker deployments without shared state or database.