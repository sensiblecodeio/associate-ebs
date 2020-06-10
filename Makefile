build:
	docker build -t associate-ebs .
	docker run --rm associate-ebs cat /go/bin/associate-ebs > associate-ebs
	chmod u+x associate-ebs

clean:
	rm associate-ebs

.PHONY: clean
