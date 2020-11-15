using System;
using System.Collections.Generic;
using System.Text;

namespace DotNet.Models {
    public class Uris {
        private readonly string _baseUrl;

        // TODO: Do we need to ensure trailing slash is/isnot present?
        public Uris(string baseUrl) {
            _baseUrl = baseUrl;
        }

        public string FetchPartitionUrl => _baseUrl + "/partitions";
    }
}
