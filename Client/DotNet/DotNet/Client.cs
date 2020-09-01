using System;
using System.Collections.Generic;
using DotNet.Models;

namespace DotNet {
    class Client {
        public void Upload(DataEntry entry) {

        }

        /// <summary>
        /// Delete all the items in a given partition
        /// The callee is responsible for filtering out duplicates which may exist locally.
        /// </summary>
        /// <param name="partition"></param>
        public void Download(string partition) {

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
        public object GetPartitions() {
            return null;
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
