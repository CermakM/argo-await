FROM scratch
COPY dist/kube-await /bin/
ENTRYPOINT [ "/bin/kube-await" ]
