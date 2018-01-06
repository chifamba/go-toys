FROM scratch
ADD ../../repo/amd64/backend-stub /bin/backend-stub
ENTRYPOINT /bin/backend-stub