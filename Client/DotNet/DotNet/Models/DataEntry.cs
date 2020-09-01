namespace DotNet.Models {
    class DataEntry {
        public EntryIdentifier Identifier { get; set; }

        /// <summary>
        /// The data to be backed up, serialized to JSON prior to upload.
        /// Callee is responsible for serialization to JSON
        /// </summary>
        public string Json { get; set; }
    }
}
