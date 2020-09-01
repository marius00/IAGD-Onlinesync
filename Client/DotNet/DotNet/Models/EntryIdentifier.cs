namespace DotNet.Models {
    class EntryIdentifier {
        /// <summary>
        /// Partition entry for this entry.
        /// The active partitions are defined by the server
        /// </summary>
        public string Partition { get; set; }

        /// <summary>
        /// Unique persistent identifier for this entry, defined by the callee.
        /// Typically Guid.NewGuid().ToString(), persisted to a local database.
        /// </summary>
        public string Id { get; set; }
    }
}
