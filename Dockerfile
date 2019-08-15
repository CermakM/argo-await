FROM scratch
COPY dist/argo-await /bin/
ENTRYPOINT [ "/bin/argo-await" ]
