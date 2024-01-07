FROM  amazoncorretto:21
RUN mkdir -p /app
COPY dist /app
RUN chmod +x /app/startsmppsim.sh
EXPOSE 8884
EXPOSE 2775
WORKDIR /app
CMD ["/app/startsmppsim.sh"]
