namespace DotNet.Models {
    class Partition {
        /// <summary>
        /// Partition entry for this entry.
        /// The active partitions are defined by the server
        /// 
        /// Not to be confused with the client provided "Id" for uploads.
        /// </summary>
        public string Id { get; set; }

        /// <summary>
        /// If this is an active or archived partition
        /// Archived partitions may be edited server-side, but no further action is required by the client.
        /// 
        /// If an item is deleted from an archived partition, it will be included as a deletion entry in the currently active partition.
        /// </summary>
        public bool IsActive { get; set; }
    }
}
