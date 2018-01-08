FROM ubuntu
ADD ./backend-stub /backend-stub
EXPOSE 8080
CMD ["/backend-stub"]