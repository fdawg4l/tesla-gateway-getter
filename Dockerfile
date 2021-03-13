FROM scratch
COPY bin/tesla-gateway-getter /tesla-gateway-getter

ENTRYPOINT ["/tesla-gateway-getter"]
CMD ["tesla-gateway-getter"]
