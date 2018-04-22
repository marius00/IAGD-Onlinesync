using System;
using System.Collections.Generic;
using System.Text;
using Microsoft.WindowsAzure.Storage.Table;

namespace ItemSync.Shared.Model
{
    public class EmailAuthToken : TableEntity {
        public const string TableName = "emailauthtoken";

        public EmailAuthToken() { }
        public EmailAuthToken(string email, string token, string code) {
            this.PartitionKey = token;
            this.RowKey = code;
            this.Email = email;
        }

        public DateTimeOffset Expiration { get; set; }
        public string Email { get; set; }
    }
}
