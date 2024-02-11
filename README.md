# GFS_MapReduce
This is a simple implementation of Google File System architecture beside MapReduce Example ,Fault Tolerance and GO Routines 
MapReduce is intended to count Number of Nulceo Bases A T C G from Genome.fasta File you can change this code from slaves with any desired function 
Note :Any Slave could be Mapper or Reducer but in this code Slave 5 is the Reducer and slaves from 1 to 4 is Mappers but Slave 1 Will work as Reducer if SLave 5 was down although it will do mapping anyway
### Don't forget to  watch the video or read the power point file
## Steps to Run The Project
0- Open 7 different terminals on VSCode IDE

1- Run salves from 1 to 5 each on different terminal tabs

2- Run divide_chunks_on_slaves from master.go file 

3- Comment divide_chunks_on_slaves function and uncomment master as server function then run master.go file again

4-There is two types of Get request in client.go file 
	First : to get fasta chunks from slaves 
  	Second :to run map function on any chunk in slaves from 1 to 4 or run MapReduce if the id = 0 on all Slaves 
	 
5- Choose desired chunk number by changing id value from the  get request at client.go file then run client.go

Note : id=0 will write all chunks after gathering them from slaves for first get type  otherwise it will run map reduce to count number of nucleo basese in Genome.fasta File
if you wants to run the code on one device only change IP address from all files to your IP
When connecting with other devices change ip addresses from master and all slave files don't change port number only this part of IPs 192.168.1.5

Example :
change ** http://192.168.1.184:8091 to ** http://192.168.43.207:8091


This Steps for first time running after that you can ignore step 2 because you had already made the chunks and 
divided them over slaves so no need to run it again except if you deleted slaves.fasta files aka chunks


### References

1- [How to make Http server in go (Go tutorial in Arabic)](https://www.youtube.com/watch?v=3mfdXvsBwdk&list=PLpxoM9UzFN77FpoSegTEKPZ512YG3rdQv&index=32)
	
 
2- [What is Map Reduce](https://www.youtube.com/watch?v=MAJ0aW5g17c&list=PLAXUYU7PbJhjoqdE-SGbBIl2jU-XBtZpP&index=2)
	
 
3- [What is a REST API (to understand the use of post and get requests)](https://www.youtube.com/watch?v=lsMQRaeKNDk)
	


