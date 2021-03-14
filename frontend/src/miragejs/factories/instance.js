import { Factory } from "miragejs";

export default {
  instance: Factory.extend({
    name(i) {
      if (i == 0) {
        return "dev env";
      } else if (i == 1) {
        return "test env";
      } else if (i == 2) {
        return "staging env";
      } else {
        return "prod env " + i;
      }
    },
    externalLink() {
      return "google.com";
    },
    host(i) {
      if (i == 0) {
        return "localhost";
      } else if (i == 1) {
        return "127.0.0.1";
      } else if (i == 2) {
        return "13.24.32.122";
      } else {
        return "mydb.com";
      }
    },
    port(i) {
      if (i == 0) {
        return "3306";
      } else if (i == 1) {
        return "";
      } else if (i == 2) {
        return "15202";
      } else {
        return "5432";
      }
    },
    afterCreate(instance, server) {
      server.create("dataSource", {
        instance,
        name: instance.name + " admin ds1",
        type: "ADMIN",
      });

      for (let i = 0; i < 3; i++) {
        const database = server.create("database", {
          instance,
          name: instance.name + " db1",
        });

        server.create("dataSource", {
          instance,
          database,
          name: instance.name + " " + " admin ds2",
          type: "ADMIN",
          username: "root",
          password: "pwdRoot",
        });

        server.create("dataSource", {
          instance,
          database,
          name: instance.name + " " + " rw ds3",
          type: "RW",
          username: "rootRW",
          password: "pwdRW",
        });

        server.create("dataSource", {
          instance,
          database,
          name: instance.name + " " + " ro ds4",
          type: "RO",
          username: "rootRO",
          password: "pwdRO",
        });
      }
    },
  }),
};
