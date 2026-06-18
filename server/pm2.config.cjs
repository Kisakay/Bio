module.exports = {
  apps: [
    {
      name: "kisakay-api",
      script: "./server",

      instances: 1,
      exec_mode: "fork",

      autorestart: true,
      watch: false,

      env: {
        NODE_ENV: "production",
      },

      time: true,
      merge_logs: true,
    },
  ],
};