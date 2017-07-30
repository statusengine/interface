<?php
/**
 * Statusengine UI
 * Copyright (C) 2016-2017  Daniel Ziegler
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

namespace Statusengine\Loader\Mysql;

use Statusengine\Backend\StorageBackend;
use Statusengine\Loader\HostDowntimeLoaderInterface;
use Statusengine\ValueObjects\HostDowntimeQueryOptions;

class HostDowntimeLoader implements HostDowntimeLoaderInterface {

    /**
     * @var \Statusengine\Backend\Mysql\MySQL
     */
    private $Backend;

    /**
     * HostDowntimeLoader constructor.
     * @param StorageBackend $StorageBackend
     */
    public function __construct(StorageBackend $StorageBackend) {
        $this->Backend = $StorageBackend->getBackend();
    }

    public function getHostdowntime(HostDowntimeQueryOptions $HostDowntimeQueryOptions) {
        $fields = [
            'HDH.internal_downtime_id',
            'HDH.scheduled_start_time',
            'HDH.scheduled_end_time',
            'HDH.author_name',
            'HDH.comment_data',
            'HDH.is_fixed',
            'HDH.node_name',

            'HSD.hostname',
            'HSD.actual_start_time'

        ];
        $sql = "SELECT %s FROM statusengine_host_downtimehistory AS HDH
             INNER JOIN statusengine_host_scheduleddowntimes AS HSD ON
                 HDH.internal_downtime_id = HSD.internal_downtime_id
              WHERE HSD.actual_start_time > 0
              AND HSD.hostname=?
              ORDER BY HDH.scheduled_start_time DESC
             LIMIT 1";

        $sql = sprintf($sql, implode(', ', $fields));

        $query = $this->Backend->prepare($sql);
        $query->bindValue(1, $HostDowntimeQueryOptions->getHostname());
        $result = $this->Backend->fetchAll($query);
        if (!empty($result)) {
            return $result;
        }
        return [];
    }
}
