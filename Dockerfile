FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-azure-devops"]
COPY baton-azure-devops /