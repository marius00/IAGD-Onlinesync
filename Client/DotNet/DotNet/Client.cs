using System;
using System.Collections.Generic;
using DotNet.Internal;
using DotNet.Models;

namespace DotNet {
    class Client {
        private readonly RestService _restService;
        private readonly Uris _uris;

        public Client(Uris uris, IJsonSerializer serializer, string authToken) {
            _restService = new RestService(serializer, authToken);
            _uris = uris;
        }

        public void Upload<T>(T entry) where T : EntryIdentifier {

        }

        public void Upload(DataEntry entry) {

        }

        /// <summary>
        /// Delete all the items in a given partition
        /// The callee is responsible for filtering out duplicates which may exist locally.
        /// </summary>
        /// <param name="partition"></param>
        public List<T> Download<T>(string partition) {
            return _restService.Get<List<T>>(_uris.FetchPartitionUrl); // TODO totally wrong.

        }

        /// <summary>
        /// Delete all the provided entries.
        /// Server side may provide a soft or hard delete, depending on configuration.
        /// </summary>
        /// <param name="entries">Collection of unique identifiers to delete remotely</param>
        public void Delete(ICollection<EntryIdentifier> entries) {

        }

        /// <summary>
        /// Delete the provided entry.
        /// Server side may provide a soft or hard delete, depending on configuration.
        /// </summary>
        /// <param name="entry">Unique identifier to delete remotely</param>
        public void Delete(EntryIdentifier entry) {
            Delete(new List<EntryIdentifier> {entry});
        }

        /// <summary>
        /// Get a list of partitions stored remotely
        /// Typically used to diff with partitions existing locally, and download new entries.
        /// </summary>
        /// <returns></returns>
        public List<Partition> GetPartitions() {
            return _restService.Get<List<Partition>>(_uris.FetchPartitionUrl);
        }
        
        /// <summary>
        /// Permanently delete all remotely stored data associated with this account/login
        /// </summary>
        /// <returns></returns>
        public bool DeleteAccount() {
            return false;
        }
    }
}
