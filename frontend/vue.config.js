module.exports = {
    devServer: {
        port: 8000,
        client: {
            webSocketURL: {
                protocol: 'wss', // Use wss for secure WebSocket connection
                hostname: 'localhost',
                port: 2443,
                pathname: '/ws',
            },
        },
    }
}