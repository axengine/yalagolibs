# cubist signer client
- Use signer-session-psbt.json, do not use signer-session.json (no PSBT signature permission by default) 
- If the signer-session-psbt.json is within the refresh validity period, you can refresh it directly 
- If the signer-session-psbt.json is not within the refresh validity period, it will fail and you can delete it 
- If the signer-session-psbt.json doesn't exist, you need to log in and use the management-session.json to generate signer-session-0.json 
- Doc：[https://signer-docs.cubist.dev/api.html](https://signer-docs.cubist.dev/api.html) 